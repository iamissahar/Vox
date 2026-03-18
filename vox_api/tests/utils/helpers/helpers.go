package helpers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
	"vox/internal/admin/logs"
	"vox/internal/auth"
	"vox/internal/hub"
	"vox/internal/user"
	"vox/internal/user/voice"
	"vox/tests/utils/mocks"
	"vox/tests/utils/vars"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

func HappyAuthDB(u auth.UserInfo) *mocks.AuthDB {
	return &mocks.AuthDB{
		GetUserF: func(_ context.Context, _ *zap.Logger, _ int, _ string) (auth.UserInfo, error) {
			return u, nil
		},
		AddNewProviderUserF: func(_ context.Context, _ *zap.Logger, _ auth.UserInfo) error {
			return nil
		},
		SaveRefreshTokenF: func(_ context.Context, _ *zap.Logger, _, _ string) error {
			return nil
		},
	}
}

func ErrorAuthDB() *mocks.AuthDB {
	dbErr := errors.New("db error")
	return &mocks.AuthDB{
		AddNewManualUserF: func(_ context.Context, _ *zap.Logger, _ auth.UserInfo, _ []byte) error {
			return dbErr
		},
		GetUserF: func(_ context.Context, _ *zap.Logger, _ int, _ string) (auth.UserInfo, error) {
			return auth.UserInfo{}, dbErr
		},
		AddNewProviderUserF: func(_ context.Context, _ *zap.Logger, _ auth.UserInfo) error {
			return dbErr
		},
		GetPasswordHashF: func(_ context.Context, _ *zap.Logger, _ string) ([]byte, error) {
			return nil, dbErr
		},
		SaveRefreshTokenF: func(_ context.Context, _ *zap.Logger, _, _ string) error {
			return dbErr
		},
		GetRefreshTokenF: func(_ context.Context, _ *zap.Logger, _ string) (string, error) {
			return "", dbErr
		},
	}
}

func BcryptHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	return string(hash)
}

func InjectLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("logger", logger)
		ctx.Next()
	}
}

func NewLoginRouter(t *testing.T, api *auth.AuthAPI) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	r.GET("/auth/:provider/login", api.LoginViaProviderHandler)
	return r
}

func NewCallbackRouter(t *testing.T, api *auth.AuthAPI) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	r.GET("/auth/:provider/callback", api.ProviderCallbackHandler)
	return r
}

func NewLoginHandlerRouter(t *testing.T, api *auth.AuthAPI) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	r.POST("/auth/login", api.LoginHandler)
	return r
}

func CallbackRequest(provider, code string) *http.Request {
	target := "/auth/" + provider + "/callback?code=" + url.QueryEscape(code)
	return httptest.NewRequest(http.MethodGet, target, nil)
}

func LoginRequest(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func OauthTokenServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
		assert.NoError(t, err)
	}))
}

func OauthTokenServerFailing(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, `{"error":"invalid_grant"}`)
		assert.NoError(t, err)
	}))
}

func UserInfoServer(t *testing.T, body map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(body)
		assert.NoError(t, err)
	}))
}

func NewRefreshRouter(t *testing.T, api *auth.AuthAPI) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	r.POST("/auth/refresh", api.RefreshHandler)
	return r
}

func RefreshRequest(accessToken, refreshToken string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	if accessToken != "" {
		req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	}
	if refreshToken != "" {
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	}
	return req
}

func ValidTokenPair(t *testing.T, subject string) (accessToken, refreshToken string) {
	t.Helper()
	cfg := vars.BaseConfig("")
	access, refresh, err := GeneratePair(zaptest.NewLogger(t), subject, cfg.JWTSecret)
	require.NoError(t, err)
	return access, refresh
}

func RefreshHashOf(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func InjectHub(h *hub.Hub) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("hub", h)
		ctx.Next()
	}
}

func NewListenRouter(t *testing.T, api *hub.HubAPI, h *hub.Hub) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	r.Use(InjectHub(h))
	r.GET("/hub/:hub_id/listen", api.ListenHandler)
	return r
}

func NewListenRouterNoHub(t *testing.T, api *hub.HubAPI) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	// No hub injected — simulates missing/invalid context value.
	r.GET("/hub/:hub_id/listen", api.ListenHandler)
	return r
}

func ListenRequest(hubID string) *http.Request {
	return httptest.NewRequest(http.MethodGet, "/hub/"+hubID+"/listen", nil)
}

func ListenRequestWithContext(hubID string, ctx context.Context) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/hub/"+hubID+"/listen", nil)
	return req.WithContext(ctx)
}

func ValidTokenPairWithSecret(t *testing.T, subject, secret string) (string, string) {
	t.Helper()
	access, refresh, err := GeneratePair(zaptest.NewLogger(t), subject, secret)
	require.NoError(t, err)
	return access, refresh
}

type claims struct {
	Issuer   string `json:"iss"`
	Audience string `json:"aud"`
	Subject  string `json:"sub"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
	Nbf      int64  `json:"nbf"`
	Jti      string `json:"jti"`
	jwt.RegisteredClaims
}

func GeneratePair(log *zap.Logger, userID, secret string) (access, refresh string, err error) {
	log.Debug("GeneratePair", zap.String("userID", userID), zap.Int("secret_length", len(secret)))
	now := time.Now().Unix()
	key := []byte(secret)

	claims := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "vox_api",
			Audience:  jwt.ClaimStrings{"admin"},
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Unix(now, 0).Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Unix(now, 0)),
			NotBefore: jwt.NewNumericDate(time.Unix(now, 0)),
			ID:        uuid.New().String(),
		},
	}

	access, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
	if err != nil {
		log.Error("Failed to sign access token", zap.Error(err))
		return
	}

	randomBytes := make([]byte, 32)
	if _, err = rand.Read(randomBytes); err != nil {
		log.Error("Failed to generate refresh token", zap.Error(err))
		return
	}
	refresh = hex.EncodeToString(randomBytes)

	log.Debug("Pair generated", zap.Bool("access_is_empty", access == ""), zap.Bool("refresh_is_empty", refresh == ""), zap.String("user_id", userID))
	return
}

func NewHubRouter(t *testing.T, api *hub.HubAPI, cache *hub.HostAndHubs, userID string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	r.Use(func(c *gin.Context) {
		c.Set("host_and_hub_cache", cache)
		if userID != "" {
			c.Set("user_id", userID)
		}
		c.Next()
	})
	r.POST("/hub", api.NewHubHandler)
	return r
}

func NewHubRequest() *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/hub", nil)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func NewReconnectRouter(t *testing.T, api *hub.HubAPI, cache *hub.HostAndHubs, userID string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	r.Use(func(c *gin.Context) {
		c.Set("host_and_hub_cache", cache)
		if userID != "" {
			c.Set("user_id", userID)
		}
		c.Next()
	})
	r.GET("/hub/:hub_id/reconnect", api.ReconnectHandler)
	return r
}

func NewReconnectRequest(hubID, userID string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/hub/"+hubID+"/reconnect", nil)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func NewDeleteHubRouter(t *testing.T, api *hub.HubAPI, cache *hub.HostAndHubs, h *hub.Hub, userID string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	r.Use(func(c *gin.Context) {
		c.Set("host_and_hub_cache", cache)
		if h != nil {
			c.Set("hub", h)
		}
		if userID != "" {
			c.Set("user_id", userID)
		}
		c.Next()
	})
	r.DELETE("/hub/:hub_id", api.DeleteHubHandler)
	return r
}

func NewDeleteHubRequest(hubID, userID string) *http.Request {
	req := httptest.NewRequest(http.MethodDelete, "/hub/"+hubID, nil)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func NewHubsRouter(t *testing.T, api *user.UserAPI, cache *hub.HostAndHubs, userID string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	r.Use(func(c *gin.Context) {
		c.Set("host_and_hub_cache", cache)
		if userID != "" {
			c.Set("user_id", userID)
		}
		c.Next()
	})
	r.GET("/user/hubs", api.HubsHandler)
	return r
}

func NewHubsRequest() *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/user/hubs", nil)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func NewPublishRouterNoUserID(t *testing.T, api *hub.HubAPI, h *hub.Hub) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	r.Use(InjectHub(h))
	r.POST("/hub/:hub_id/publish", api.PublishHandler)
	return r
}

func NewPublishRouterNoHub(t *testing.T, api *hub.HubAPI, userID string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", userID)
		ctx.Next()
	})
	r.POST("/hub/:hub_id/publish", api.PublishHandler)
	return r
}

func PublishRequest(hubID string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/hub/"+hubID+"/publish", nil)
}

func WriteTempFile(t *testing.T, content []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "ref-*.mp3")
	require.NoError(t, err)
	_, err = f.Write(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func closeWebSocket(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	err := conn.Close()
	assert.NoError(t, err)
}

func NewMockDeepgramServer(t *testing.T, transcript string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := vars.WsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer closeWebSocket(t, conn)

		if transcript != "" {
			msg := map[string]interface{}{
				"type": "Results",
				"channel": map[string]interface{}{
					"alternatives": []map[string]interface{}{
						{"transcript": transcript},
					},
				},
				"is_final": true,
			}
			data, _ := json.Marshal(msg)
			_ = conn.WriteMessage(websocket.TextMessage, data)
		}

		_ = conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func NewMockGroqServer(t *testing.T, content string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "chat/completions") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)

		flusher := w.(http.Flusher)

		chunk := map[string]interface{}{
			"id":     "chatcmpl-test",
			"object": "chat.completion.chunk",
			"model":  "llama3-8b-8192",
			"choices": []map[string]interface{}{
				{"index": 0, "delta": map[string]string{"content": content}},
			},
		}
		data, _ := json.Marshal(chunk)
		_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
		flusher.Flush()

		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	t.Cleanup(srv.Close)
	return srv
}

func NewMockFishAudioServer(t *testing.T, audioChunk []byte) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := vars.WsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("WsUpgrader.Upgrade error: %v", err)
			return
		}
		defer closeWebSocket(t, conn)

		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()

		time.Sleep(10 * time.Millisecond)

		if len(audioChunk) > 0 {
			_ = conn.WriteMessage(websocket.BinaryMessage, audioChunk)
		}
		_ = conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		time.Sleep(30 * time.Millisecond)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func NewPublishRouterFull(t *testing.T, api *hub.HubAPI, h *hub.Hub, userID string, fishBuilder hub.FishBuilder) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	r.Use(InjectHub(h))
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", userID)
		ctx.Set("fish_builder", fishBuilder)
		ctx.Next()
	})
	r.POST("/hub/:hub_id/publish", api.PublishHandler)
	return r
}

func NewConsumer(h *hub.Hub) (*hub.Consumer, <-chan []byte) {
	c := &hub.Consumer{
		ID:   uuid.New().String(),
		Send: make(chan []byte, 32),
	}
	h.AddConsumer(c)
	return c, c.Send
}

func HappyHubDB(filename, text string) *mocks.HubDB {
	return &mocks.HubDB{
		GetReferenceF: func(_ context.Context, _ *zap.Logger, _ string) (string, string, error) {
			return filename, text, nil
		},
	}
}

func ErrorHubDB() *mocks.HubDB {
	return &mocks.HubDB{
		GetReferenceF: func(_ context.Context, _ *zap.Logger, _ string) (string, string, error) {
			return "", "", errors.New("db error")
		},
	}
}

func NewInfoHandlerRouter(t *testing.T, api *user.UserAPI, userID any) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	r.Use(func(ctx *gin.Context) {
		if userID != nil {
			ctx.Set("user_id", userID)
		}
		ctx.Next()
	})
	r.GET("/user/info", api.InfoHandler)
	return r
}

func NewReferenceHandlerRouter(t *testing.T, api *voice.VoiceAPI, userID any) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	r.Use(func(ctx *gin.Context) {
		if userID != nil {
			ctx.Set("user_id", userID)
		}
		ctx.Next()
	})
	r.POST("/voice/reference", api.ReferenceHandler)
	return r
}

type CloseNotifyRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func NewCloseNotifyRecorder() *CloseNotifyRecorder {
	return &CloseNotifyRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		closed:           make(chan bool, 1),
	}
}

func (c *CloseNotifyRecorder) CloseNotify() <-chan bool {
	return c.closed
}

func Argon2Hash(password string) []byte {
	salt := make([]byte, 16)
	rand.Read(salt)
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return []byte(hex.EncodeToString(hash) + "$" + hex.EncodeToString(salt))
}

func InsertRefreshToken(t *testing.T, userID string, refreshToken string, db *pgxpool.Pool) {
	_, err := db.Exec(context.Background(), "INSERT INTO auth (user_id, refresh_token) VALUES ($1, $2)", userID, refreshToken)
	require.NoError(t, err)
}

func InsertPasswordHash(t *testing.T, userID string, passwordHash []byte, db *pgxpool.Pool) {
	_, err := db.Exec(context.Background(), "INSERT INTO auth_references (user_id, password_hash) VALUES ($1, $2)", userID, passwordHash)
	require.NoError(t, err)
}

func InsertVoiceRef(t *testing.T, userID string, filename string, text string, db *pgxpool.Pool) {
	_, err := db.Exec(context.Background(), "INSERT INTO user_voice (user_id, filename, text) VALUES ($1, $2, $3)", userID, filename, text)
	require.NoError(t, err)
}

func InsertAdditionalUserInfo(t *testing.T, user vars.UserForTests, db *pgxpool.Pool) {
	_, err := db.Exec(context.Background(), "INSERT INTO users (id, email, name, picture_url) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET email = $2, name = $3, picture_url = $4", user.ID, user.Email, user.Name, user.Picture)
	require.NoError(t, err)
}

func InsertFileMetadata(t *testing.T, fileID string, path string, typeof string, db *pgxpool.Pool) {
	_, err := db.Exec(context.Background(), "INSERT INTO files (id, full_path, type, text) VALUES ($1, $2, $3, $4)", fileID, path, typeof, "")
	require.NoError(t, err)
}

func InsertProviderUserRef(t *testing.T, userID string, providerID int, userProviderID string, db *pgxpool.Pool) {
	_, err := db.Exec(context.Background(), "INSERT INTO users_and_providers (user_id, provider_id, user_provider_id) VALUES ($1, $2, $3)", userID, providerID, userProviderID)
	require.NoError(t, err)
}

func NewLevelRouter(t *testing.T, api *logs.LogsAPI) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InjectLogger(zaptest.NewLogger(t)))
	r.PUT("/logs/level", api.LevelHandler)
	return r
}

func NewLogsAPI() (*logs.LogsAPI, zap.AtomicLevel) {
	atom := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	return &logs.LogsAPI{Atomic: atom}, atom
}

package vars

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"vox/internal/auth"
	mod "vox/pkg/models"

	"github.com/gorilla/websocket"
)

var GoogleUserInfo = map[string]any{
	"id":      "google-uid-123",
	"email":   "user@example.com",
	"name":    "Test User",
	"picture": "https://example.com/pic.jpg",
}
var GithubUserInfo = map[string]any{
	"id":         99999999,
	"email":      "user@example.com",
	"name":       "Test User",
	"avatar_url": "https://example.com/avatar.jpg",
}
var ExistingUser = auth.UserInfo{
	ID:    "user-uuid-1",
	Email: "user@example.com",
	Name:  "Test User",
}
var WsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsURL(srv *httptest.Server) string {
	return "ws" + strings.TrimPrefix(srv.URL, "http")
}

func CfgWithMocks(dgSrv, groqSrv, fishSrv *httptest.Server) *mod.Config {
	return &mod.Config{
		DeepgramAPIKey:  "test-key",
		DeepgramBaseURL: wsURL(dgSrv),
		DeepgramModel:   "nova-2",
		GroqAPIKey:      "test-key",
		GroqModel:       "llama3-8b-8192",
		GroqBaseURL:     groqSrv.URL + "/openai/v1",
	}
}

func PublishCfg() *mod.Config {
	return &mod.Config{
		DeepgramAPIKey:   "invalid-key",
		DeepgramBaseURL:  "https://127.0.0.1:0",
		DeepgramModel:    "nova-2",
		GroqAPIKey:       "invalid-key",
		GroqModel:        "llama3-8b-8192",
		GroqBaseURL:      "https://127.0.0.1:0",
		FishAudioAPIKey:  "invalid-key",
		FishAudioBaseURL: "https://127.0.0.1:0",
	}
}

func BaseConfig(tokenURL string) *mod.Config {
	return &mod.Config{
		BaseURL:            "https://example.com",
		FrontendURL:        "https://frontend.example.com",
		JWTSecret:          "test-secret-that-is-long-enough-32b",
		GoogleClientID:     "test-google-client-id",
		GoogleClientSecret: "test-google-client-secret",
		GithubClientID:     "test-github-client-id",
		GithubClientSecret: "test-github-client-secret",
		GoogleTokenURL:     tokenURL + "/token",
		GithubTokenURL:     tokenURL + "/token",
		StoragePath:        os.TempDir() + "/",
	}
}

type UserForTests struct {
	ID             string
	Email          string
	Name           string
	Picture        string
	ProviderID     int
	UserProviderID string
}

var User UserForTests = UserForTests{
	ID:             "user-123",
	Email:          "alice@example.com",
	Name:           "Alice",
	Picture:        "https://example.com/pic.jpg",
	ProviderID:     -1,
	UserProviderID: "google-123",
}
var Hash []byte = []byte("somehashedbytes")
var RefreshToken = "somerefreshtoken"
var PasswordHash []byte = []byte("somepasswordhash")

const GOOGLE_PROVIDER_ID = -1

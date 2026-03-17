//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"vox/internal/hub"
	"vox/tests/utils/helpers"
	"vox/tests/utils/mocks"
	"vox/tests/utils/vars"
)

func TestPublishHandler_MissingUserID_Returns404(t *testing.T) {
	h := hub.NewHub(uuid.New().String())
	defer h.Close()

	api := &hub.HubAPI{MGR: hub.NewManager(), Cfg: vars.PublishCfg(), DB: helpers.HappyHubDB("", "")}
	r := helpers.NewPublishRouterNoUserID(t, api, h)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(h.ID))

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPublishHandler_MissingHub_Returns404(t *testing.T) {
	api := &hub.HubAPI{MGR: hub.NewManager(), Cfg: vars.PublishCfg(), DB: helpers.HappyHubDB("", "")}
	r := helpers.NewPublishRouterNoHub(t, api, uuid.New().String())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(uuid.New().String()))

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPublishHandler_WrongHubType_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(helpers.InjectLogger(zaptest.NewLogger(t)))
	r.Use(func(ctx *gin.Context) {
		ctx.Set("hub", "not-a-hub-pointer")
		ctx.Set("user_id", uuid.New().String())
		ctx.Next()
	})
	api := &hub.HubAPI{MGR: hub.NewManager(), Cfg: vars.PublishCfg(), DB: helpers.HappyHubDB("", "")}
	r.POST("/hub/:hub_id/publish", api.PublishHandler)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(uuid.New().String()))

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPublishHandler_DBError_Returns500(t *testing.T) {
	h := hub.NewHub(uuid.New().String())
	defer h.Close()

	api := &hub.HubAPI{MGR: hub.NewManager(), Cfg: vars.PublishCfg(), DB: helpers.ErrorHubDB()}
	r := helpers.NewPublishRouterFull(t, api, h, uuid.New().String(), mocks.HappyFishBuilder())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(h.ID))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPublishHandler_DBError_ResponseBodyIsValidJSON(t *testing.T) {
	h := hub.NewHub(uuid.New().String())
	defer h.Close()

	api := &hub.HubAPI{MGR: hub.NewManager(), Cfg: vars.PublishCfg(), DB: helpers.ErrorHubDB()}
	r := helpers.NewPublishRouterFull(t, api, h, uuid.New().String(), mocks.HappyFishBuilder())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(h.ID))

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.True(t, json.Valid(w.Body.Bytes()))
}

func TestPublishHandler_FileNotFound_Returns500(t *testing.T) {
	h := hub.NewHub(uuid.New().String())
	defer h.Close()

	api := &hub.HubAPI{
		MGR: hub.NewManager(),
		Cfg: vars.PublishCfg(),
		DB:  helpers.HappyHubDB("/nonexistent/does-not-exist.mp3", "text"),
	}
	r := helpers.NewPublishRouterFull(t, api, h, uuid.New().String(), mocks.HappyFishBuilder())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(h.ID))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPublishHandler_FileNotFound_ResponseBodyIsValidJSON(t *testing.T) {
	h := hub.NewHub(uuid.New().String())
	defer h.Close()

	api := &hub.HubAPI{
		MGR: hub.NewManager(),
		Cfg: vars.PublishCfg(),
		DB:  helpers.HappyHubDB("/nonexistent/does-not-exist.mp3", "text"),
	}
	r := helpers.NewPublishRouterFull(t, api, h, uuid.New().String(), mocks.HappyFishBuilder())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(h.ID))

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.True(t, json.Valid(w.Body.Bytes()))
}

// defer MGR.Delete is registered AFTER the DB+file reads, so on a DB error
// the hub must still be present in the manager when the handler returns.
func TestPublishHandler_DBError_HubNotRemovedFromManager(t *testing.T) {
	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)

	api := &hub.HubAPI{MGR: mgr, Cfg: vars.PublishCfg(), DB: helpers.ErrorHubDB()}
	r := helpers.NewPublishRouterFull(t, api, h, uuid.New().String(), mocks.HappyFishBuilder())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(hubID))

	require.Equal(t, http.StatusInternalServerError, w.Code)
	_, ok := mgr.Get(hubID)
	assert.True(t, ok, "hub must still exist: defer Delete not registered at DB-error return point")
}

// Same for file-not-found.
func TestPublishHandler_FileNotFound_HubNotRemovedFromManager(t *testing.T) {
	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)

	api := &hub.HubAPI{
		MGR: mgr,
		Cfg: vars.PublishCfg(),
		DB:  helpers.HappyHubDB("/nonexistent/does-not-exist.mp3", "text"),
	}
	r := helpers.NewPublishRouterFull(t, api, h, uuid.New().String(), mocks.HappyFishBuilder())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(hubID))

	require.Equal(t, http.StatusInternalServerError, w.Code)
	_, ok := mgr.Get(hubID)
	assert.True(t, ok, "hub must still exist: defer Delete not registered at file-error return point")
}
func TestPublishHandler_HappyPath_Returns200(t *testing.T) {
	refFile := helpers.WriteTempFile(t, []byte("fake-reference-audio"))
	dgSrv := helpers.NewMockDeepgramServer(t, "hello world")
	groqSrv := helpers.NewMockGroqServer(t, "processed token")
	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)
	api := &hub.HubAPI{
		MGR: mgr,
		Cfg: vars.CfgWithMocks(dgSrv, groqSrv, nil),
		DB:  helpers.HappyHubDB(refFile, "reference text"),
	}
	r := helpers.NewPublishRouterFull(t, api, h, uuid.New().String(), mocks.HappyFishBuilder())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(hubID))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPublishHandler_FishError_Returns500(t *testing.T) {
	refFile := helpers.WriteTempFile(t, []byte("fake-reference-audio"))
	dgSrv := helpers.NewMockDeepgramServer(t, "hello world")
	groqSrv := helpers.NewMockGroqServer(t, "processed token")
	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)
	api := &hub.HubAPI{
		MGR: mgr,
		Cfg: vars.CfgWithMocks(dgSrv, groqSrv, nil),
		DB:  helpers.HappyHubDB(refFile, "reference text"),
	}
	r := helpers.NewPublishRouterFull(t, api, h, uuid.New().String(), mocks.ErrorFishBuilder())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(hubID))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPublishHandler_NoFishBuilder_Returns500(t *testing.T) {
	refFile := helpers.WriteTempFile(t, []byte("fake-reference-audio"))
	dgSrv := helpers.NewMockDeepgramServer(t, "hello world")
	groqSrv := helpers.NewMockGroqServer(t, "processed token")
	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)
	api := &hub.HubAPI{
		MGR: mgr,
		Cfg: vars.CfgWithMocks(dgSrv, groqSrv, nil),
		DB:  helpers.HappyHubDB(refFile, "reference text"),
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(helpers.InjectLogger(zaptest.NewLogger(t)))
	r.Use(helpers.InjectHub(h))
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uuid.New().String())
		ctx.Next()
	})
	r.POST("/hub/:hub_id/publish", api.PublishHandler)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(hubID))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPublishHandler_InvalidFishBuilder_Returns500(t *testing.T) {
	refFile := helpers.WriteTempFile(t, []byte("fake-reference-audio"))
	dgSrv := helpers.NewMockDeepgramServer(t, "hello world")
	groqSrv := helpers.NewMockGroqServer(t, "processed token")
	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)
	api := &hub.HubAPI{
		MGR: mgr,
		Cfg: vars.CfgWithMocks(dgSrv, groqSrv, nil),
		DB:  helpers.HappyHubDB(refFile, "reference text"),
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(helpers.InjectLogger(zaptest.NewLogger(t)))
	r.Use(helpers.InjectHub(h))
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uuid.New().String())
		ctx.Set("fish_builder", "not-a-fish-builder")
		ctx.Next()
	})
	r.POST("/hub/:hub_id/publish", api.PublishHandler)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(hubID))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPublishHandler_HappyPath_HubRemovedFromManagerAfterReturn(t *testing.T) {
	refFile := helpers.WriteTempFile(t, []byte("fake-reference-audio"))

	dgSrv := helpers.NewMockDeepgramServer(t, "hello world")
	groqSrv := helpers.NewMockGroqServer(t, "token")
	fishSrv := helpers.NewMockFishAudioServer(t, []byte("chunk"))

	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)

	api := &hub.HubAPI{
		MGR: mgr,
		Cfg: vars.CfgWithMocks(dgSrv, groqSrv, fishSrv),
		DB:  helpers.HappyHubDB(refFile, "reference text"),
	}
	r := helpers.NewPublishRouterFull(t, api, h, uuid.New().String(), mocks.HappyFishBuilder())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(hubID))

	require.Equal(t, http.StatusOK, w.Code)
	_, ok := mgr.Get(hubID)
	assert.False(t, ok, "defer MGR.Delete must have fired: hub must be gone after handler returns")
}

func TestPublishHandler_HappyPath_AudioChunkReachesHubConsumer(t *testing.T) {
	refFile := helpers.WriteTempFile(t, []byte("fake-reference-audio"))

	const expectedAudio = "fish-audio-bytes"

	dgSrv := helpers.NewMockDeepgramServer(t, "hello")
	groqSrv := helpers.NewMockGroqServer(t, "token")
	fishSrv := helpers.NewMockFishAudioServer(t, []byte(expectedAudio))

	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)

	consumer, ch := helpers.NewConsumer(h)
	defer h.RemoveConsumer(consumer.ID)

	received := make(chan []byte, 16)
	go func() {
		for chunk := range ch {
			received <- chunk
		}
	}()

	api := &hub.HubAPI{
		MGR: mgr,
		Cfg: vars.CfgWithMocks(dgSrv, groqSrv, fishSrv),
		DB:  helpers.HappyHubDB(refFile, "reference text"),
	}
	r := helpers.NewPublishRouterFull(t, api, h, uuid.New().String(), mocks.HappyFishBuilder())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(hubID))

	require.Equal(t, http.StatusOK, w.Code)

	select {
	case chunk := <-received:
		assert.Equal(t, []byte(expectedAudio), chunk)
	case <-time.After(2 * time.Second):
		t.Fatal("audio chunk was never delivered to the hub consumer")
	}
}

func TestPublishHandler_DeepgramUnreachable_Returns500(t *testing.T) {
	refFile := helpers.WriteTempFile(t, []byte("fake-reference-audio"))

	// Start a server and close it immediately so the address is unreachable.
	deadSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadSrv.Close()

	groqSrv := helpers.NewMockGroqServer(t, "token")
	fishSrv := helpers.NewMockFishAudioServer(t, []byte("chunk"))

	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)

	api := &hub.HubAPI{
		MGR: mgr,
		Cfg: vars.CfgWithMocks(deadSrv, groqSrv, fishSrv),
		DB:  helpers.HappyHubDB(refFile, "reference text"),
	}
	r := helpers.NewPublishRouterFull(t, api, h, uuid.New().String(), mocks.HappyFishBuilder())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(hubID))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPublishHandler_GroqUnreachable_Returns500(t *testing.T) {
	refFile := helpers.WriteTempFile(t, []byte("fake-reference-audio"))

	dgSrv := helpers.NewMockDeepgramServer(t, "hello world")

	deadSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadSrv.Close()

	fishSrv := helpers.NewMockFishAudioServer(t, []byte("chunk"))

	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)

	api := &hub.HubAPI{
		MGR: mgr,
		Cfg: vars.CfgWithMocks(dgSrv, deadSrv, fishSrv),
		DB:  helpers.HappyHubDB(refFile, "reference text"),
	}
	r := helpers.NewPublishRouterFull(t, api, h, uuid.New().String(), mocks.HappyFishBuilder())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(hubID))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPublishHandler_FishAudioUnreachable_Returns500(t *testing.T) {
	refFile := helpers.WriteTempFile(t, []byte("fake-reference-audio"))

	dgSrv := helpers.NewMockDeepgramServer(t, "hello world")
	groqSrv := helpers.NewMockGroqServer(t, "token")

	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)

	api := &hub.HubAPI{
		MGR: mgr,
		Cfg: vars.CfgWithMocks(dgSrv, groqSrv, nil),
		DB:  helpers.HappyHubDB(refFile, "reference text"),
	}
	r := helpers.NewPublishRouterFull(t, api, h, uuid.New().String(), mocks.ErrorFishBuilder())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.PublishRequest(hubID))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPublishHandler_HappyPath_DoesNotPanic(t *testing.T) {
	refFile := helpers.WriteTempFile(t, []byte("fake-reference-audio"))

	dgSrv := helpers.NewMockDeepgramServer(t, "hello")
	groqSrv := helpers.NewMockGroqServer(t, "token")
	fishSrv := helpers.NewMockFishAudioServer(t, []byte("chunk"))

	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)

	api := &hub.HubAPI{
		MGR: mgr,
		Cfg: vars.CfgWithMocks(dgSrv, groqSrv, fishSrv),
		DB:  helpers.HappyHubDB(refFile, "reference text"),
	}
	r := helpers.NewPublishRouterFull(t, api, h, uuid.New().String(), mocks.HappyFishBuilder())

	assert.NotPanics(t, func() {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, helpers.PublishRequest(hubID))
	})
}

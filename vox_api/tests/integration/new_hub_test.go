//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"vox/internal/hub"
	"vox/tests/utils/helpers"
)

func TestNewHubHandler_StatusCreated(t *testing.T) {
	cache := hub.NewHostAndHubs()
	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewHubRouter(t, api, cache)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubRequest(uuid.New().String()))

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestNewHubHandler_ResponseIsValidJSON(t *testing.T) {
	cache := hub.NewHostAndHubs()
	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewHubRouter(t, api, cache)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubRequest(uuid.New().String()))

	require.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, json.Valid(w.Body.Bytes()))
}

func TestNewHubHandler_ResponseContainsNonEmptyHubID(t *testing.T) {
	cache := hub.NewHostAndHubs()
	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewHubRouter(t, api, cache)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubRequest(uuid.New().String()))

	require.Equal(t, http.StatusCreated, w.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.NotEmpty(t, body["hub_id"])
}

func TestNewHubHandler_ResponseHubIDIsValidUUID(t *testing.T) {
	cache := hub.NewHostAndHubs()
	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewHubRouter(t, api, cache)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubRequest(uuid.New().String()))

	require.Equal(t, http.StatusCreated, w.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	_, err := uuid.Parse(body["hub_id"])
	assert.NoError(t, err)
}

func TestNewHubHandler_HubRegisteredInCache(t *testing.T) {
	cache := hub.NewHostAndHubs()
	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewHubRouter(t, api, cache)

	userID := uuid.New().String()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubRequest(userID))

	require.Equal(t, http.StatusCreated, w.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	hubs := cache.GetHubs(userID)
	assert.Contains(t, hubs, body["hub_id"])
}

func TestNewHubHandler_ContentTypeIsJSON(t *testing.T) {
	cache := hub.NewHostAndHubs()
	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewHubRouter(t, api, cache)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubRequest(uuid.New().String()))

	require.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestNewHubHandler_InvalidBody_ReturnsBadRequest(t *testing.T) {
	cache := hub.NewHostAndHubs()
	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewHubRouter(t, api, cache)

	req := httptest.NewRequest(http.MethodPost, "/hub", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNewHubHandler_MissingCache_ReturnsInternalError(t *testing.T) {
	// роутер без инжекта кэша
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(helpers.InjectLogger(zaptest.NewLogger(t)))
	r.POST("/hub", (&hub.HubAPI{MGR: hub.NewManager()}).NewHubHandler)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubRequest(uuid.New().String()))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNewHubHandler_MultipleUsers_CacheIsolated(t *testing.T) {
	// хабы разных пользователей не пересекаются в кэше
	cache := hub.NewHostAndHubs()
	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewHubRouter(t, api, cache)

	userA := uuid.New().String()
	userB := uuid.New().String()

	wA := httptest.NewRecorder()
	r.ServeHTTP(wA, helpers.NewHubRequest(userA))
	require.Equal(t, http.StatusCreated, wA.Code)

	wB := httptest.NewRecorder()
	r.ServeHTTP(wB, helpers.NewHubRequest(userB))
	require.Equal(t, http.StatusCreated, wB.Code)

	var bodyA, bodyB map[string]string
	require.NoError(t, json.Unmarshal(wA.Body.Bytes(), &bodyA))
	require.NoError(t, json.Unmarshal(wB.Body.Bytes(), &bodyB))

	assert.NotContains(t, cache.GetHubs(userA), bodyB["hub_id"])
	assert.NotContains(t, cache.GetHubs(userB), bodyA["hub_id"])
}

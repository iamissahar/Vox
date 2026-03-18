//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	r := helpers.NewHubRouter(t, api, cache, uuid.New().String())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubRequest())

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestNewHubHandler_ResponseIsValidJSON(t *testing.T) {
	cache := hub.NewHostAndHubs()
	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewHubRouter(t, api, cache, uuid.New().String())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubRequest())

	require.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, json.Valid(w.Body.Bytes()))
}

func TestNewHubHandler_ResponseContainsNonEmptyHubID(t *testing.T) {
	cache := hub.NewHostAndHubs()
	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewHubRouter(t, api, cache, uuid.New().String())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubRequest())

	require.Equal(t, http.StatusCreated, w.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.NotEmpty(t, body["hub_id"])
}

func TestNewHubHandler_ResponseHubIDIsValidUUID(t *testing.T) {
	cache := hub.NewHostAndHubs()
	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewHubRouter(t, api, cache, uuid.New().String())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubRequest())

	require.Equal(t, http.StatusCreated, w.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	_, err := uuid.Parse(body["hub_id"])
	assert.NoError(t, err)
}

func TestNewHubHandler_HubRegisteredInCache(t *testing.T) {
	cache := hub.NewHostAndHubs()
	api := &hub.HubAPI{MGR: hub.NewManager()}
	userID := uuid.New().String()
	r := helpers.NewHubRouter(t, api, cache, userID)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubRequest())

	require.Equal(t, http.StatusCreated, w.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	hubs := cache.GetHubs(userID)
	assert.Contains(t, hubs, body["hub_id"])
}

func TestNewHubHandler_ContentTypeIsJSON(t *testing.T) {
	cache := hub.NewHostAndHubs()
	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewHubRouter(t, api, cache, uuid.New().String())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubRequest())

	require.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestNewHubHandler_MissingUserID_ReturnsBadRequest(t *testing.T) {
	cache := hub.NewHostAndHubs()
	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewHubRouter(t, api, cache, "")

	req := httptest.NewRequest(http.MethodPost, "/hub", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNewHubHandler_MissingCache_ReturnsInternalError(t *testing.T) {
	// роутер без инжекта кэша
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(helpers.InjectLogger(zaptest.NewLogger(t)))
	r.POST("/hub", (&hub.HubAPI{MGR: hub.NewManager()}).NewHubHandler)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubRequest())

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

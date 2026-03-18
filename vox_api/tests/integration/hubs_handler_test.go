//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"vox/internal/hub"
	"vox/internal/user"
	"vox/tests/utils/helpers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestHubsHandler_HappyPath_ReturnsOK(t *testing.T) {
	cache := hub.NewHostAndHubs()
	userID := uuid.New().String()
	hubID := uuid.New().String()
	cache.AddHub(userID, hubID)

	r := helpers.NewHubsRouter(t, &user.UserAPI{}, cache, userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubsRequest())

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHubsHandler_HappyPath_ResponseIsValidJSON(t *testing.T) {
	cache := hub.NewHostAndHubs()
	userID := uuid.New().String()
	cache.AddHub(userID, uuid.New().String())

	r := helpers.NewHubsRouter(t, &user.UserAPI{}, cache, userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubsRequest())

	require.Equal(t, http.StatusOK, w.Code)
	assert.True(t, json.Valid(w.Body.Bytes()))
}

func TestHubsHandler_HappyPath_ResponseContainsHubID(t *testing.T) {
	cache := hub.NewHostAndHubs()
	userID := uuid.New().String()
	hubID := uuid.New().String()
	cache.AddHub(userID, hubID)

	r := helpers.NewHubsRouter(t, &user.UserAPI{}, cache, userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubsRequest())

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), hubID)
}

func TestHubsHandler_HappyPath_MultipleHubs(t *testing.T) {
	cache := hub.NewHostAndHubs()
	userID := uuid.New().String()
	hubID1 := uuid.New().String()
	hubID2 := uuid.New().String()
	cache.AddHub(userID, hubID1)
	cache.AddHub(userID, hubID2)

	r := helpers.NewHubsRouter(t, &user.UserAPI{}, cache, userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubsRequest())

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), hubID1)
	assert.Contains(t, w.Body.String(), hubID2)
}

func TestHubsHandler_HappyPath_NoHubs_ReturnsEmptyArray(t *testing.T) {
	cache := hub.NewHostAndHubs()
	userID := uuid.New().String()

	r := helpers.NewHubsRouter(t, &user.UserAPI{}, cache, userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubsRequest())

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"hub_ids"`)
}

func TestHubsHandler_HappyPath_ContentTypeIsJSON(t *testing.T) {
	cache := hub.NewHostAndHubs()
	userID := uuid.New().String()
	cache.AddHub(userID, uuid.New().String())

	r := helpers.NewHubsRouter(t, &user.UserAPI{}, cache, userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubsRequest())

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestHubsHandler_HappyPath_OtherUserHubsNotReturned(t *testing.T) {
	cache := hub.NewHostAndHubs()
	userID := uuid.New().String()
	otherUserID := uuid.New().String()
	otherHubID := uuid.New().String()
	cache.AddHub(otherUserID, otherHubID)

	r := helpers.NewHubsRouter(t, &user.UserAPI{}, cache, userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubsRequest())

	require.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), otherHubID)
}

func TestHubsHandler_MissingCache_ReturnsInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(helpers.InjectLogger(zaptest.NewLogger(t)))
	r.GET("/user/hubs", (&user.UserAPI{}).HubsHandler)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubsRequest())

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHubsHandler_MissingUserID_ReturnsInternalError(t *testing.T) {
	cache := hub.NewHostAndHubs()

	r := helpers.NewHubsRouter(t, &user.UserAPI{}, cache, "")
	req := httptest.NewRequest(http.MethodGet, "/user/hubs", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHubsHandler_InvalidCacheType_ReturnsInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(helpers.InjectLogger(zaptest.NewLogger(t)))
	r.Use(func(c *gin.Context) {
		c.Set("host_and_hub_cache", "not a cache")
		c.Set("user_id", "str")
		c.Next()
	})
	r.GET("/user/hubs", (&user.UserAPI{}).HubsHandler)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewHubsRequest())

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

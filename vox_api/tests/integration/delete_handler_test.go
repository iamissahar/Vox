//go:build integration

package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"vox/internal/hub"
	"vox/tests/utils/helpers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestDeleteHubHandler_HappyPath_ReturnsNoContent(t *testing.T) {
	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)

	cache := hub.NewHostAndHubs()
	userID := uuid.New().String()
	cache.AddHub(userID, hubID)

	api := &hub.HubAPI{MGR: mgr}
	r := helpers.NewDeleteHubRouter(t, api, cache, h, userID)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewDeleteHubRequest(hubID, userID))

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteHubHandler_HappyPath_HubRemovedFromManager(t *testing.T) {
	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)

	cache := hub.NewHostAndHubs()
	userID := uuid.New().String()
	cache.AddHub(userID, hubID)

	api := &hub.HubAPI{MGR: mgr}
	r := helpers.NewDeleteHubRouter(t, api, cache, h, userID)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewDeleteHubRequest(hubID, userID))

	require.Equal(t, http.StatusNoContent, w.Code)
	_, ok := mgr.Get(hubID)
	assert.False(t, ok, "hub must be removed from manager after delete")
}

func TestDeleteHubHandler_NotOwner_ReturnsForbidden(t *testing.T) {
	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)

	cache := hub.NewHostAndHubs()
	userID := uuid.New().String()
	cache.AddHub(uuid.New().String(), hubID)

	api := &hub.HubAPI{MGR: mgr}
	r := helpers.NewDeleteHubRouter(t, api, cache, h, userID)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewDeleteHubRequest(hubID, uuid.New().String()))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteHubHandler_NotOwner_HubStillExistsInManager(t *testing.T) {
	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)

	cache := hub.NewHostAndHubs()
	userID := uuid.New().String()
	cache.AddHub(uuid.New().String(), hubID)

	api := &hub.HubAPI{MGR: mgr}
	r := helpers.NewDeleteHubRouter(t, api, cache, h, userID)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewDeleteHubRequest(hubID, uuid.New().String()))

	require.Equal(t, http.StatusForbidden, w.Code)
	_, ok := mgr.Get(hubID)
	assert.True(t, ok, "hub must still exist after forbidden delete attempt")
}

func TestDeleteHubHandler_MissingUserID_ReturnsInternalError(t *testing.T) {
	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)

	cache := hub.NewHostAndHubs()
	cache.AddHub(uuid.New().String(), hubID)

	api := &hub.HubAPI{MGR: mgr}
	r := helpers.NewDeleteHubRouter(t, api, cache, h, "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewDeleteHubRequest(hubID, ""))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteHubHandler_MissingHub_ReturnsNotFound(t *testing.T) {
	mgr := hub.NewManager()
	cache := hub.NewHostAndHubs()
	api := &hub.HubAPI{MGR: mgr}

	userID := uuid.New().String()

	// hub не инжектится в контекст
	r := helpers.NewDeleteHubRouter(t, api, cache, nil, userID)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewDeleteHubRequest(uuid.New().String(), uuid.New().String()))

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteHubHandler_MissingCache_ReturnsInternalError(t *testing.T) {
	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)
	userID := uuid.New().String()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(helpers.InjectLogger(zaptest.NewLogger(t)))
	r.Use(func(c *gin.Context) {
		c.Set("hub", h) // кэш не инжектим
		c.Set("user_id", userID)
		c.Next()
	})
	r.DELETE("/hub/:hub_id", (&hub.HubAPI{MGR: mgr}).DeleteHubHandler)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewDeleteHubRequest(hubID, uuid.New().String()))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteHubHandler_InvalidCacheType_ReturnsInternalError(t *testing.T) {
	mgr := hub.NewManager()
	hubID := mgr.New()
	h, _ := mgr.Get(hubID)
	userID := uuid.New().String()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(helpers.InjectLogger(zaptest.NewLogger(t)))
	r.Use(func(c *gin.Context) {
		c.Set("hub", h)
		c.Set("host_and_hub_cache", "not a cache")
		c.Set("user_id", userID)
		c.Next()
	})
	r.DELETE("/hub/:hub_id", (&hub.HubAPI{MGR: mgr}).DeleteHubHandler)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewDeleteHubRequest(hubID, uuid.New().String()))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteHubHandler_UserWithMultipleHubs_DeletesOnlyTargetHub(t *testing.T) {
	mgr := hub.NewManager()
	hubID1 := mgr.New()
	hubID2 := mgr.New()
	h1, _ := mgr.Get(hubID1)

	cache := hub.NewHostAndHubs()
	userID := uuid.New().String()
	cache.AddHub(userID, hubID1)
	cache.AddHub(userID, hubID2)

	api := &hub.HubAPI{MGR: mgr}
	r := helpers.NewDeleteHubRouter(t, api, cache, h1, userID)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewDeleteHubRequest(hubID1, userID))

	require.Equal(t, http.StatusNoContent, w.Code)

	_, ok1 := mgr.Get(hubID1)
	assert.False(t, ok1, "hub1 must be deleted")

	_, ok2 := mgr.Get(hubID2)
	assert.True(t, ok2, "hub2 must still exist")
}

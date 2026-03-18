//go:build integration

package integration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"vox/internal/hub"
	"vox/pkg/models"
	"vox/tests/utils/helpers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestReconnectHandler_HappyPath_RedirectsToPublish(t *testing.T) {
	cache := hub.NewHostAndHubs()
	cfg := &models.Config{FrontendURL: "https://example.com"}
	api := &hub.HubAPI{MGR: hub.NewManager(), Cfg: cfg}

	userID := uuid.New().String()
	hubID := uuid.New().String()
	cache.AddHub(userID, hubID)

	r := helpers.NewReconnectRouter(t, api, cache)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewReconnectRequest(hubID, userID))

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, cfg.FrontendURL+"/hub/"+hubID+"/publish", w.Header().Get("Location"))
}

func TestReconnectHandler_NotOwner_ReturnsForbidden(t *testing.T) {
	cache := hub.NewHostAndHubs()
	cfg := &models.Config{FrontendURL: "https://example.com"}
	api := &hub.HubAPI{MGR: hub.NewManager(), Cfg: cfg}

	// другой юзер владеет хабом
	cache.AddHub(uuid.New().String(), uuid.New().String())

	r := helpers.NewReconnectRouter(t, api, cache)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewReconnectRequest(uuid.New().String(), uuid.New().String()))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestReconnectHandler_InvalidBody_ReturnsBadRequest(t *testing.T) {
	cache := hub.NewHostAndHubs()
	cfg := &models.Config{FrontendURL: "https://example.com"}
	api := &hub.HubAPI{MGR: hub.NewManager(), Cfg: cfg}

	req := httptest.NewRequest(http.MethodGet, "/hub/"+uuid.New().String()+"/reconnect", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")

	r := helpers.NewReconnectRouter(t, api, cache)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReconnectHandler_EmptyBody_ReturnsBadRequest(t *testing.T) {
	cache := hub.NewHostAndHubs()
	cfg := &models.Config{FrontendURL: "https://example.com"}
	api := &hub.HubAPI{MGR: hub.NewManager(), Cfg: cfg}

	req := httptest.NewRequest(http.MethodGet, "/hub/"+uuid.New().String()+"/reconnect", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

	r := helpers.NewReconnectRouter(t, api, cache)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// user_id пустой — юзер не владелец ни одного хаба
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestReconnectHandler_MissingCache_ReturnsInternalError(t *testing.T) {
	cfg := &models.Config{FrontendURL: "https://example.com"}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(helpers.InjectLogger(zaptest.NewLogger(t)))
	r.GET("/hub/:hub_id/reconnect", (&hub.HubAPI{MGR: hub.NewManager(), Cfg: cfg}).ReconnectHandler)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewReconnectRequest(uuid.New().String(), uuid.New().String()))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestReconnectHandler_UserWithMultipleHubs_CorrectHubReconnects(t *testing.T) {
	cache := hub.NewHostAndHubs()
	cfg := &models.Config{FrontendURL: "https://example.com"}
	api := &hub.HubAPI{MGR: hub.NewManager(), Cfg: cfg}

	userID := uuid.New().String()
	hubID1 := uuid.New().String()
	hubID2 := uuid.New().String()
	cache.AddHub(userID, hubID1)
	cache.AddHub(userID, hubID2)

	r := helpers.NewReconnectRouter(t, api, cache)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewReconnectRequest(hubID1, userID))
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, cfg.FrontendURL+"/hub/"+hubID1+"/publish", w.Header().Get("Location"))

	w = httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewReconnectRequest(hubID2, userID))
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, cfg.FrontendURL+"/hub/"+hubID2+"/publish", w.Header().Get("Location"))
}

func TestReconnectHandler_InvalidCacheType_ReturnsInternalError(t *testing.T) {
	cfg := &models.Config{FrontendURL: "https://example.com"}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(helpers.InjectLogger(zaptest.NewLogger(t)))
	r.Use(func(c *gin.Context) {
		c.Set("host_and_hub_cache", "not a cache")
		c.Next()
	})
	r.GET("/hub/:hub_id/reconnect", (&hub.HubAPI{MGR: hub.NewManager(), Cfg: cfg}).ReconnectHandler)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.NewReconnectRequest(uuid.New().String(), uuid.New().String()))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

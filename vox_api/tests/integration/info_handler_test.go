//go:build integration

package integration

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"vox/internal/user"
	"vox/tests/utils/helpers"
	"vox/tests/utils/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInfoHandler_UserIDNotSet(t *testing.T) {
	api := &user.UserAPI{DB: &mocks.UserDB{}}
	r := helpers.NewInfoHandlerRouter(t, api, nil) // no user_id injected
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/info", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInfoHandler_UserIDWrongType(t *testing.T) {
	api := &user.UserAPI{DB: &mocks.UserDB{}}
	r := helpers.NewInfoHandlerRouter(t, api, 12345) // int, not string
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/info", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInfoHandler_GetUserInfoDBError(t *testing.T) {
	db := &mocks.UserDB{
		GetUserInfoF: func(_ context.Context, _ *zap.Logger, _ string) (user.UserInfo, error) {
			return user.UserInfo{}, errors.New("db error")
		},
	}
	api := &user.UserAPI{DB: db}
	r := helpers.NewInfoHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/info", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestInfoHandler_HappyPath(t *testing.T) {
	expected := user.UserInfo{
		ID:      "user-123",
		Email:   "user@example.com",
		Name:    "John Doe",
		Picture: "https://example.com/pic.jpg",
	}
	db := &mocks.UserDB{
		GetUserInfoF: func(_ context.Context, _ *zap.Logger, _ string) (user.UserInfo, error) {
			return expected, nil
		},
	}
	api := &user.UserAPI{DB: db}
	r := helpers.NewInfoHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/info", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{
		"id":      "user-123",
		"email":   "user@example.com",
		"name":    "John Doe",
		"picture": "https://example.com/pic.jpg"
	}`, w.Body.String())
}

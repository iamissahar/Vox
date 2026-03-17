//go:build integration

package integration

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"vox/internal/auth"
	"vox/tests/utils/helpers"
	"vox/tests/utils/mocks"
	"vox/tests/utils/vars"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestRefreshHandler_MissingAccessTokenCookie(t *testing.T) {
	api := &auth.AuthAPI{DB: &mocks.AuthDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewRefreshRouter(t, api)

	// No cookies at all.
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRefreshHandler_MissingRefreshTokenCookie(t *testing.T) {
	api := &auth.AuthAPI{DB: &mocks.AuthDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewRefreshRouter(t, api)

	// Only access_token cookie, no refresh_token.
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "some-token"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRefreshHandler_InvalidAccessToken(t *testing.T) {
	api := &auth.AuthAPI{DB: &mocks.AuthDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewRefreshRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.RefreshRequest("not-a-valid-jwt", "some-refresh-token"))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRefreshHandler_AccessTokenSignedWithWrongSecret(t *testing.T) {
	// Generate a token with a different secret than what the API uses.
	access, _, err := helpers.GeneratePair(zaptest.NewLogger(t), "user-id-1", "a-completely-different-secret")
	require.NoError(t, err)

	api := &auth.AuthAPI{DB: &mocks.AuthDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewRefreshRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.RefreshRequest(access, "any-refresh"))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRefreshHandler_ExpiredAccessTokenStillAccepted(t *testing.T) {
	access, refresh := helpers.ValidTokenPair(t, "user-expired-1")
	hash := helpers.RefreshHashOf(refresh)

	db := &mocks.AuthDB{
		GetRefreshTokenF: func(_ context.Context, _ *zap.Logger, _ string) (string, error) {
			return hash, nil
		},
		SaveRefreshTokenF: func(_ context.Context, _ *zap.Logger, _, _ string) error {
			return nil
		},
	}
	api := &auth.AuthAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewRefreshRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.RefreshRequest(access, refresh))

	// 201 expected — expiry must not block the refresh path.
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRefreshHandler_GetRefreshTokenDBError(t *testing.T) {
	access, refresh := helpers.ValidTokenPair(t, "user-db-err-1")

	db := &mocks.AuthDB{
		GetRefreshTokenF: func(_ context.Context, _ *zap.Logger, _ string) (string, error) {
			return "", errors.New("db error")
		},
	}
	api := &auth.AuthAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewRefreshRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.RefreshRequest(access, refresh))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRefreshHandler_RefreshTokenHashMismatch(t *testing.T) {
	access, _ := helpers.ValidTokenPair(t, "user-mismatch-1")

	db := &mocks.AuthDB{
		GetRefreshTokenF: func(_ context.Context, _ *zap.Logger, _ string) (string, error) {
			// Return a hash that will never match the presented token.
			return helpers.RefreshHashOf("completely-different-token"), nil
		},
	}
	api := &auth.AuthAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewRefreshRouter(t, api)

	w := httptest.NewRecorder()
	// Present a refresh token whose hash doesn't match the stored one.
	r.ServeHTTP(w, helpers.RefreshRequest(access, "tampered-refresh-token"))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefreshHandler_GeneratePairFails(t *testing.T) {
	cfg := vars.BaseConfig("")
	cfg.JWTSecret = ""
	access, refresh := helpers.ValidTokenPairWithSecret(t, "user-genpair-fail-1", cfg.JWTSecret)
	hash := helpers.RefreshHashOf(refresh)

	db := &mocks.AuthDB{
		GetRefreshTokenF: func(_ context.Context, _ *zap.Logger, _ string) (string, error) {
			return hash, nil
		},
		SaveRefreshTokenF: func(_ context.Context, _ *zap.Logger, _, _ string) error {
			return nil
		},
	}
	api := &auth.AuthAPI{DB: db, Cfg: cfg}
	r := helpers.NewRefreshRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.RefreshRequest(access, refresh))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRefreshHandler_SaveRefreshTokenDBError(t *testing.T) {
	access, refresh := helpers.ValidTokenPair(t, "user-save-fail-1")
	hash := helpers.RefreshHashOf(refresh)

	db := &mocks.AuthDB{
		GetRefreshTokenF: func(_ context.Context, _ *zap.Logger, _ string) (string, error) {
			return hash, nil
		},
		SaveRefreshTokenF: func(_ context.Context, _ *zap.Logger, _, _ string) error {
			return errors.New("db error")
		},
	}
	api := &auth.AuthAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewRefreshRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.RefreshRequest(access, refresh))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRefreshHandler_HappyPath(t *testing.T) {
	access, refresh := helpers.ValidTokenPair(t, "user-happy-refresh-1")
	hash := helpers.RefreshHashOf(refresh)

	var savedHash string
	db := &mocks.AuthDB{
		GetRefreshTokenF: func(_ context.Context, _ *zap.Logger, _ string) (string, error) {
			return hash, nil
		},
		SaveRefreshTokenF: func(_ context.Context, _ *zap.Logger, _, h string) error {
			savedHash = h
			return nil
		},
	}
	api := &auth.AuthAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewRefreshRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.RefreshRequest(access, refresh))

	require.Equal(t, http.StatusCreated, w.Code)

	// Both token cookies must be present and non-empty.
	cookies := map[string]*http.Cookie{}
	for _, c := range w.Result().Cookies() {
		cookies[c.Name] = c
	}
	require.Contains(t, cookies, "access_token")
	require.Contains(t, cookies, "refresh_token")
	assert.NotEmpty(t, cookies["access_token"].Value)
	assert.NotEmpty(t, cookies["refresh_token"].Value)

	// The newly issued tokens must differ from the originals (rotation).
	assert.NotEqual(t, access, cookies["access_token"].Value)
	assert.NotEqual(t, refresh, cookies["refresh_token"].Value)

	// The hash persisted to the DB must match the new refresh token's hash.
	assert.Equal(t, helpers.RefreshHashOf(cookies["refresh_token"].Value), savedHash)
}

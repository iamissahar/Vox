//go:build integration

package integration

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/iotest"
	"vox/internal/auth"
	"vox/tests/utils/helpers"
	"vox/tests/utils/mocks"
	"vox/tests/utils/vars"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestLoginHandler_BodyUnreadable(t *testing.T) {
	api := &auth.AuthAPI{DB: &mocks.AuthDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewLoginHandlerRouter(t, api)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", iotest.ErrReader(errors.New("read error")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginHandler_BodyNotJSON(t *testing.T) {
	api := &auth.AuthAPI{DB: &mocks.AuthDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewLoginHandlerRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.LoginRequest(`not-json{{`))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginHandler_GetPasswordHashDBError(t *testing.T) {
	db := &mocks.AuthDB{
		GetPasswordHashF: func(_ context.Context, _ *zap.Logger, _ string) ([]byte, error) {
			return nil, errors.New("db error")
		},
	}
	api := &auth.AuthAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewLoginHandlerRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.LoginRequest(`{"login":"user@example.com","password":"anypassword"}`))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLoginHandler_WrongPassword(t *testing.T) {
	db := &mocks.AuthDB{
		GetPasswordHashF: func(_ context.Context, _ *zap.Logger, _ string) ([]byte, error) {
			return []byte("different-password"), nil
		},
	}
	api := &auth.AuthAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewLoginHandlerRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.LoginRequest(`{"login":"user@example.com","password":"wrongpassword"}`))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLoginHandler_GeneratePairFails(t *testing.T) {
	db := &mocks.AuthDB{
		GetPasswordHashF: func(_ context.Context, _ *zap.Logger, _ string) ([]byte, error) {
			return nil, errors.New("db error")
		},
	}
	cfg := vars.BaseConfig("")
	cfg.JWTSecret = "" // empty secret causes generatePair to fail
	api := &auth.AuthAPI{DB: db, Cfg: cfg}
	r := helpers.NewLoginHandlerRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.LoginRequest(`{"login":"user@example.com","password":"correctpassword"}`))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLoginHandler_SaveRefreshTokenDBError(t *testing.T) {
	db := &mocks.AuthDB{
		GetPasswordHashF: func(_ context.Context, _ *zap.Logger, _ string) ([]byte, error) {
			return nil, errors.New("db error")
		},
		SaveRefreshTokenF: func(_ context.Context, _ *zap.Logger, _, _ string) error {
			return errors.New("db error")
		},
	}
	api := &auth.AuthAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewLoginHandlerRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.LoginRequest(`{"login":"user@example.com","password":"correctpassword"}`))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLoginHandler_HappyPath(t *testing.T) {
	db := &mocks.AuthDB{
		GetPasswordHashF: func(_ context.Context, _ *zap.Logger, _ string) ([]byte, error) {
			return helpers.Argon2Hash("correctpassword"), nil
		},
		SaveRefreshTokenF: func(_ context.Context, _ *zap.Logger, _, _ string) error {
			return nil
		},
	}

	api := &auth.AuthAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewLoginHandlerRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.LoginRequest(`{"login":"user@example.com","password":"correctpassword"}`))

	require.Equal(t, http.StatusOK, w.Code)

	var cookieNames []string
	for _, c := range w.Result().Cookies() {
		cookieNames = append(cookieNames, c.Name)
	}
	assert.Contains(t, cookieNames, "access_token")
	assert.Contains(t, cookieNames, "refresh_token")
}

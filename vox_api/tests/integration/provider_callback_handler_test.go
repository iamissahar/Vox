//go:build integration

package integration

import (
	"context"
	"errors"
	"io"
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
)

func TestProviderCallbackHandler_UnsupportedProvider(t *testing.T) {
	api := &auth.AuthAPI{DB: &mocks.AuthDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewCallbackRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.CallbackRequest("facebook", "any-code"))

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestProviderCallbackHandler_ExchangeFails(t *testing.T) {
	tokenSrv := helpers.OauthTokenServerFailing(t)
	defer tokenSrv.Close()

	api := &auth.AuthAPI{DB: &mocks.AuthDB{}, Cfg: vars.BaseConfig(tokenSrv.URL)}
	r := helpers.NewCallbackRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.CallbackRequest("google", "bad-code"))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProviderCallbackHandler_UserinfoUnreachable(t *testing.T) {
	tokenSrv := helpers.OauthTokenServer(t)
	defer tokenSrv.Close()

	deadSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadSrv.Close()

	cfg := vars.BaseConfig(tokenSrv.URL)
	cfg.GoogleUserInfoURL = deadSrv.URL + "/userinfo"

	api := &auth.AuthAPI{DB: &mocks.AuthDB{}, Cfg: cfg}
	r := helpers.NewCallbackRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.CallbackRequest("google", "valid-code"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestProviderCallbackHandler_UserinfoMalformedBody(t *testing.T) {
	tokenSrv := helpers.OauthTokenServer(t)
	defer tokenSrv.Close()

	badSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `not-valid-json{{{{`)
		require.NoError(t, err)
	}))
	defer badSrv.Close()

	cfg := vars.BaseConfig(tokenSrv.URL)
	cfg.GoogleUserInfoURL = badSrv.URL + "/userinfo"

	api := &auth.AuthAPI{DB: &mocks.AuthDB{}, Cfg: cfg}
	r := helpers.NewCallbackRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.CallbackRequest("google", "valid-code"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestProviderCallbackHandler_GetUserDBError(t *testing.T) {
	tokenSrv := helpers.OauthTokenServer(t)
	defer tokenSrv.Close()

	uiSrv := helpers.UserInfoServer(t, vars.GoogleUserInfo)
	defer uiSrv.Close()

	cfg := vars.BaseConfig(tokenSrv.URL)
	cfg.GoogleUserInfoURL = uiSrv.URL + "/userinfo"

	api := &auth.AuthAPI{DB: helpers.ErrorAuthDB(), Cfg: cfg}
	r := helpers.NewCallbackRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.CallbackRequest("google", "valid-code"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestProviderCallbackHandler_AddNewProviderUserDBError(t *testing.T) {
	tokenSrv := helpers.OauthTokenServer(t)
	defer tokenSrv.Close()

	uiSrv := helpers.UserInfoServer(t, vars.GoogleUserInfo)
	defer uiSrv.Close()

	cfg := vars.BaseConfig(tokenSrv.URL)
	cfg.GoogleUserInfoURL = uiSrv.URL + "/userinfo"

	db := helpers.ErrorAuthDB()
	db.GetUserF = func(_ context.Context, _ *zap.Logger, _ int, _ string) (auth.UserInfo, error) {
		return auth.UserInfo{}, nil // not found — triggers AddNewProviderUser
	}

	api := &auth.AuthAPI{DB: db, Cfg: cfg}
	r := helpers.NewCallbackRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.CallbackRequest("google", "valid-code"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestProviderCallbackHandler_SaveRefreshTokenDBError(t *testing.T) {
	tokenSrv := helpers.OauthTokenServer(t)
	defer tokenSrv.Close()

	uiSrv := helpers.UserInfoServer(t, vars.GoogleUserInfo)
	defer uiSrv.Close()

	cfg := vars.BaseConfig(tokenSrv.URL)
	cfg.GoogleUserInfoURL = uiSrv.URL + "/userinfo"

	db := helpers.HappyAuthDB(vars.ExistingUser)
	db.SaveRefreshTokenF = func(_ context.Context, _ *zap.Logger, _, _ string) error {
		return errors.New("db error")
	}

	api := &auth.AuthAPI{DB: db, Cfg: cfg}
	r := helpers.NewCallbackRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.CallbackRequest("google", "valid-code"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestProviderCallbackHandler_HappyPath_ExistingUser_Google(t *testing.T) {
	tokenSrv := helpers.OauthTokenServer(t)
	defer tokenSrv.Close()

	uiSrv := helpers.UserInfoServer(t, vars.GoogleUserInfo)
	defer uiSrv.Close()

	cfg := vars.BaseConfig(tokenSrv.URL)
	cfg.GoogleUserInfoURL = uiSrv.URL + "/userinfo"

	api := &auth.AuthAPI{DB: helpers.HappyAuthDB(vars.ExistingUser), Cfg: cfg}
	r := helpers.NewCallbackRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.CallbackRequest("google", "valid-code"))

	require.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, "https://frontend.example.com", w.Header().Get("Location"))

	var cookieNames []string
	for _, c := range w.Result().Cookies() {
		cookieNames = append(cookieNames, c.Name)
	}
	assert.Contains(t, cookieNames, "access_token")
	assert.Contains(t, cookieNames, "refresh_token")
}

func TestProviderCallbackHandler_HappyPath_NewUser_GitHub(t *testing.T) {
	tokenSrv := helpers.OauthTokenServer(t)
	defer tokenSrv.Close()

	uiSrv := helpers.UserInfoServer(t, vars.GithubUserInfo)
	defer uiSrv.Close()

	cfg := vars.BaseConfig(tokenSrv.URL)
	cfg.GithubUserInfoURL = uiSrv.URL + "/user"

	db := helpers.HappyAuthDB(vars.ExistingUser)
	db.GetUserF = func(_ context.Context, _ *zap.Logger, _ int, _ string) (auth.UserInfo, error) {
		return auth.UserInfo{}, nil // not found — triggers AddNewProviderUser
	}

	api := &auth.AuthAPI{DB: db, Cfg: cfg}
	r := helpers.NewCallbackRouter(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.CallbackRequest("github", "valid-code"))

	require.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, "https://frontend.example.com", w.Header().Get("Location"))

	var cookieNames []string
	for _, c := range w.Result().Cookies() {
		cookieNames = append(cookieNames, c.Name)
	}
	assert.Contains(t, cookieNames, "access_token")
	assert.Contains(t, cookieNames, "refresh_token")
}

//go:build integration

package integration

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"vox/internal/auth"
	"vox/tests/utils/helpers"
	"vox/tests/utils/mocks"
	"vox/tests/utils/vars"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginViaProviderHandler_Google(t *testing.T) {
	api := &auth.AuthAPI{DB: &mocks.AuthDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewLoginRouter(t, api)

	req := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusTemporaryRedirect, w.Code)

	redirectURL, err := url.Parse(w.Header().Get("Location"))
	require.NoError(t, err)

	assert.Equal(t, "accounts.google.com", redirectURL.Host)
	assert.Equal(t, "/o/oauth2/auth", redirectURL.Path)
	assert.Equal(t, "test-google-client-id", redirectURL.Query().Get("client_id"))
	assert.Equal(t, "https://example.com/auth/google/callback", redirectURL.Query().Get("redirect_uri"))
}

func TestLoginViaProviderHandler_GitHub(t *testing.T) {
	api := &auth.AuthAPI{DB: &mocks.AuthDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewLoginRouter(t, api)

	req := httptest.NewRequest(http.MethodGet, "/auth/github/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusTemporaryRedirect, w.Code)

	redirectURL, err := url.Parse(w.Header().Get("Location"))
	require.NoError(t, err)

	assert.Equal(t, "github.com", redirectURL.Host)
	assert.Equal(t, "/login/oauth/authorize", redirectURL.Path)
	assert.Equal(t, "test-github-client-id", redirectURL.Query().Get("client_id"))
	assert.Equal(t, "https://example.com/auth/github/callback", redirectURL.Query().Get("redirect_uri"))
}

func TestLoginViaProviderHandler_UnsupportedProvider(t *testing.T) {
	api := &auth.AuthAPI{DB: &mocks.AuthDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewLoginRouter(t, api)

	req := httptest.NewRequest(http.MethodGet, "/auth/facebook/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

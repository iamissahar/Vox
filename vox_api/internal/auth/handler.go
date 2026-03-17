package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	mod "vox/pkg/models"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

func setTokenCookies(ctx *gin.Context, accessToken, refreshToken string) {
	ctx.SetCookie("access_token", accessToken, 900, "/", "", true, true)
	ctx.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", "", true, true)
}

func (a *AuthAPI) IsAdmin(ctx *gin.Context) {
	token := ctx.GetHeader("X-Admin-Token")
	if token != a.Cfg.AdminToken {
		ctx.Abort()
		ctx.Data(http.StatusUnauthorized, mod.APP_JSON, mod.HttpError(mod.UNAUTHORIZED_CODE, mod.UNAUTHORIZED_MSG))
		return
	}
	ctx.Next()
}

func (a *AuthAPI) IsAuthorized(ctx *gin.Context) {
	log := mod.GetLogger(ctx)
	cookie, err := ctx.Cookie("access_token")
	if err != nil {
		log.Warn("Access_token is missing")
		ctx.Abort()
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.INVALID_COOKIE_CODE, mod.INVALID_COOKIE_MSG))
		return
	}

	claims, err := decodeAccessToken(log, cookie, []byte(a.Cfg.JWTSecret), true)
	if err != nil {
		ctx.Abort()
		ctx.Data(http.StatusUnauthorized, mod.APP_JSON, mod.HttpError(mod.UNAUTHORIZED_CODE, mod.UNAUTHORIZED_MSG))
		return
	}

	ctx.Set("claims", claims)
	ctx.Next()
	log.Debug("User is authorized", zap.String("user_id", claims.Subject))
}

// LoginViaProviderHandler godoc
// @Summary      Login via OAuth2 provider
// @Description  Redirects the user to the authorization page of the specified OAuth2 provider (google or github)
// @Tags         auth
// @Param        provider  path  string  true  "OAuth2 provider" Enums(google, github)
// @Success      307  "Redirect to the provider authorization page"
// @Failure      404  "Provider is not supported"
// @Router       /auth/{provider}/login [get]
func (a *AuthAPI) LoginViaProviderHandler(ctx *gin.Context) {
	var (
		providerStr = ctx.Param("provider")
		redirectURL = a.Cfg.BaseURL + "/auth/" + providerStr + "/callback"
		Cfg         oauth2.Config
		log         = mod.GetLogger(ctx)
	)

	switch providerStr {
	case "google":
		log.Debug("Provides is chosen", zap.String("provider", providerStr))
		Cfg = oauth2.Config{
			ClientID:     a.Cfg.GoogleClientID,
			ClientSecret: a.Cfg.GoogleClientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}
	case "github":
		log.Debug("Provides is chosen", zap.String("provider", providerStr))
		Cfg = oauth2.Config{
			ClientID:     a.Cfg.GithubClientID,
			ClientSecret: a.Cfg.GithubClientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     github.Endpoint,
		}
	default:
		log.Warn("Provider is not supported", zap.String("provider", providerStr))
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.Redirect(http.StatusTemporaryRedirect, Cfg.AuthCodeURL("state-token"))
}

// ProviderCallbackHandler godoc
// @Summary      OAuth2 provider callback
// @Description  Handles the callback from the OAuth2 provider after user authorization. Exchanges the authorization code for a token, retrieves user info, creates the user if not exists, generates a JWT pair and sets them as cookies, then redirects to the frontend.
// @Tags         auth
// @Param        provider  path   string  true  "OAuth2 provider" Enums(google, github)
// @Param        code      query  string  true  "Authorization code returned by the provider"
// @Success      307  "Redirect to the frontend with access and refresh token cookies set"
// @Failure      401  {object}  mod.HttpErrorResponse  "Failed to exchange authorization code for token"
// @Failure      404  "Provider is not supported"
// @Failure      500  {object}  mod.HttpErrorResponse  "Internal server error"
// @Router       /auth/{provider}/callback [get]
func (a *AuthAPI) ProviderCallbackHandler(ctx *gin.Context) {
	var (
		getURL      string
		provider    Provider
		providerStr = ctx.Param("provider")
		code        = ctx.Query("code")
		redirectURL = a.Cfg.BaseURL + "/auth/" + providerStr + "/callback"
		log         = mod.GetLogger(ctx)
	)

	switch providerStr {
	case "google":
		log.Debug("Provides is chosen", zap.String("provider", providerStr))
		provider = &googleWrap{oauth2.Config{
			ClientID:     a.Cfg.GoogleClientID,
			ClientSecret: a.Cfg.GoogleClientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: oauth2.Endpoint{
				AuthURL:  google.Endpoint.AuthURL,
				TokenURL: a.Cfg.GoogleTokenURL,
			},
		}, a.DB, log}
		// https://www.googleapis.com/oauth2/v2/userinfo
		getURL = a.Cfg.GoogleUserInfoURL
	case "github":
		log.Debug("Provides is chosen", zap.String("provider", providerStr))
		provider = &githubWrap{oauth2.Config{
			ClientID:     a.Cfg.GithubClientID,
			ClientSecret: a.Cfg.GithubClientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"user:email",
				"read:user",
			},
			Endpoint: oauth2.Endpoint{
				AuthURL:  github.Endpoint.AuthURL,
				TokenURL: a.Cfg.GithubTokenURL,
			},
		}, a.DB, log}
		// https://api.github.com/user
		getURL = a.Cfg.GithubUserInfoURL
	default:
		log.Warn("Provider is not supported", zap.String("provider", providerStr))
		ctx.Status(http.StatusNotFound)
		return
	}

	token, err := provider.Exchange(ctx, code)
	if err != nil {
		ctx.Data(http.StatusUnauthorized, mod.APP_JSON, mod.HttpError(mod.UNAUTHORIZED_CODE, mod.UNAUTHORIZED_MSG))
		return
	}
	log.Debug("Code exchanged")

	resp, err := provider.Client(ctx, token).Get(getURL)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}
	log.Debug("Response received")

	user, err := provider.Read(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}
	log.Debug("Provider's data is read")

	u, ok, err := user.Get(ctx.Request.Context())
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}
	if !ok {
		u, err = user.Create(ctx.Request.Context())
		if err != nil {
			ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
			return
		}
	}

	accessToken, refreshToken, err := generatePair(log, u.ID, a.Cfg.JWTSecret)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	h := sha256.Sum256([]byte(refreshToken))
	refreshHash := hex.EncodeToString(h[:])
	if err := a.DB.SaveRefreshToken(ctx.Request.Context(), log, u.ID, refreshHash); err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	setTokenCookies(ctx, accessToken, refreshToken)
	ctx.Redirect(http.StatusTemporaryRedirect, a.Cfg.FrontendURL)
}

// SignUpHandler godoc
// @Summary      Sign up with login and password
// @Description  Registers a new user with the provided credentials, generates a JWT pair and sets them as cookies.
// @Tags         auth
// @Accept       json
// @Param        body  body  signUpPayload  true  "Sign up payload"
// @Success      201  "User created successfully, access and refresh token cookies set"
// @Failure      400  {object}  mod.HttpErrorResponse  "Invalid request body"
// @Failure      500  {object}  mod.HttpErrorResponse  "Internal server error"
// @Router       /auth/sign_up [post]
func (a *AuthAPI) SignUpHandler(ctx *gin.Context) {
	log := mod.GetLogger(ctx)
	login := signUpPayload{}
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Warn("Request body is unreadable")
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.INVALID_PAYLOAD_CODE, mod.INVALID_PAYLOAD_MSG))
		return
	}

	err = sonic.Unmarshal(body, &login)
	if err != nil {
		log.Warn("Request body is unmarshalable")
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.INVALID_PAYLOAD_CODE, mod.INVALID_PAYLOAD_MSG))
		return
	}

	hash, err := generateArgon2Hash(login.Password)
	if err != nil {
		log.Error("Failed to create password hash", zap.Error(err))
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	u := UserInfo{
		ID:             login.Login,
		Name:           login.Name,
		Email:          login.Email,
		UserProviderID: login.Login,
		ProviderID:     _MANUAL_PROVIDER_ID,
	}

	err = a.DB.AddNewManualUser(ctx.Request.Context(), log, u, hash)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	accessToken, refreshToken, err := generatePair(log, u.ID, a.Cfg.JWTSecret)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	h := sha256.Sum256([]byte(refreshToken))
	refreshHash := hex.EncodeToString(h[:])
	if err := a.DB.SaveRefreshToken(ctx.Request.Context(), log, u.ID, refreshHash); err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	setTokenCookies(ctx, accessToken, refreshToken)
	ctx.Status(http.StatusCreated)
}

// LoginHandler godoc
// @Summary      Login with login and password
// @Description  Authenticates a user with the provided credentials, generates a JWT pair and sets them as cookies.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  loginPayload  true  "Login payload"
// @Success      200  {object}  map[string]string  "Login successful, access and refresh token cookies set"
// @Failure      400  {object}  mod.HttpErrorResponse  "Invalid request body"
// @Failure      401  {object}  mod.HttpErrorResponse  "Invalid credentials"
// @Failure      500  {object}  mod.HttpErrorResponse  "Internal server error"
// @Router       /auth/login [post]
func (a *AuthAPI) LoginHandler(ctx *gin.Context) {
	log := mod.GetLogger(ctx)
	payload := new(loginPayload)
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Warn("Request body is unreadable")
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.INVALID_PAYLOAD_CODE, mod.INVALID_PAYLOAD_MSG))
		return
	}

	err = sonic.Unmarshal(body, payload)
	if err != nil {
		log.Warn("Request body is unmarshalable")
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.INVALID_PAYLOAD_CODE, mod.INVALID_PAYLOAD_MSG))
		return
	}

	passwordHash, err := a.DB.GetPasswordHash(ctx.Request.Context(), log, payload.Login)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	if err := verifyPassword(log, string(passwordHash), payload.Password); err != nil {
		ctx.Data(http.StatusUnauthorized, mod.APP_JSON, mod.HttpError(mod.UNAUTHORIZED_CODE, mod.UNAUTHORIZED_MSG))
		return
	}

	accessToken, refreshToken, err := generatePair(log, payload.Login, a.Cfg.JWTSecret)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	h := sha256.Sum256([]byte(refreshToken))
	refreshHash := hex.EncodeToString(h[:])
	if err := a.DB.SaveRefreshToken(ctx.Request.Context(), log, payload.Login, refreshHash); err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	setTokenCookies(ctx, accessToken, refreshToken)
	ctx.Data(http.StatusOK, mod.APP_JSON, []byte(`{"ok": true, status: "login"}`))
}

// RefreshHandler godoc
// @Summary      Refresh JWT token pair
// @Description  Validates the existing access and refresh token cookies, issues a new JWT pair and updates the cookies.
// @Tags         auth
// @Param        access_token   header  string  true  "Access token cookie"
// @Param        refresh_token  header  string  true  "Refresh token cookie"
// @Success      201  "Token pair refreshed successfully, new cookies set"
// @Failure      400  {object}  mod.HttpErrorResponse  "Access or refresh token cookie is missing"
// @Failure      401  {object}  mod.HttpErrorResponse  "Refresh token is invalid"
// @Failure      403  {object}  mod.HttpErrorResponse  "Access token is invalid"
// @Failure      500  {object}  mod.HttpErrorResponse  "Internal server error"
// @Router       /auth/refresh [post]
func (a *AuthAPI) RefreshHandler(ctx *gin.Context) {
	log := mod.GetLogger(ctx)
	access, err := ctx.Cookie("access_token")
	if err != nil {
		log.Warn("Access token cookie not found")
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.MISSING_COOKIE_CODE, mod.MISSING_COOKIE_MSG))
		return
	}
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		log.Warn("Refresh token cookie not found")
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.MISSING_COOKIE_CODE, mod.MISSING_COOKIE_MSG))
		return
	}

	claims, err := decodeAccessToken(log, access, []byte(a.Cfg.JWTSecret), false)
	if err != nil {
		ctx.Data(http.StatusForbidden, mod.APP_JSON, mod.HttpError(mod.INVALID_COOKIE_CODE, mod.INVALID_COOKIE_MSG))
		return
	}

	storedHash, err := a.DB.GetRefreshToken(ctx.Request.Context(), log, claims.Subject)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	if !isRefreshValid(log, refreshToken, storedHash) {
		log.Warn("Refresh token is invalid", zap.String("user_id", claims.Subject))
		ctx.Data(http.StatusUnauthorized, mod.APP_JSON, mod.HttpError(mod.UNAUTHORIZED_CODE, mod.UNAUTHORIZED_MSG))
		return
	}

	newAccess, newRefresh, err := generatePair(log, claims.Subject, a.Cfg.JWTSecret)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	h := sha256.Sum256([]byte(newRefresh))
	refreshHash := hex.EncodeToString(h[:])
	if err := a.DB.SaveRefreshToken(ctx.Request.Context(), log, claims.Subject, refreshHash); err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	setTokenCookies(ctx, newAccess, newRefresh)
	ctx.Status(http.StatusCreated)
}

package auth

import (
	"context"
	"io"
	"net/http"
	mod "vox/pkg/models"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type claims struct {
	jwt.RegisteredClaims
}

type signUpPayload struct {
	Login    string `json:"login"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginPayload struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthAPI struct {
	DB  AuthDB
	Cfg *mod.Config
}

type googleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
	Name    string `json:"name"`
	db      AuthDB
	log     *zap.Logger
}

type githubUser struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Name      string `json:"name"`
	db        AuthDB
	log       *zap.Logger
}

type UserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
	Name    string `json:"name"`

	UserProviderID string
	ProviderID     int
}

type githubWrap struct {
	oauth2.Config
	db  AuthDB
	log *zap.Logger
}

type googleWrap struct {
	oauth2.Config
	db  AuthDB
	log *zap.Logger
}

type User interface {
	Get(ctx context.Context) (u UserInfo, ok bool, err error)
	Create(ctx context.Context) (u UserInfo, err error)
}

type Provider interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, t *oauth2.Token) *http.Client
	Read(rd io.Reader) (User, error)
}

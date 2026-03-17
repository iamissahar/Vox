package models

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	INVALID_PAYLOAD_CODE      = 0
	INTERNAL_ERROR_CODE       = 1
	UNAUTHORIZED_CODE         = 2
	MISSING_COOKIE_CODE       = 3
	INVALID_COOKIE_CODE       = 4
	INVALID_CONTENT_TYPE_CODE = 5
	INVALID_URL_CODE          = 6
	ENTITY_EXISTS_CODE        = 7
)

const (
	APP_JSON                 = "application/json"
	INTERNAL_ERROR_MSG       = "something went wrong"
	UNAUTHORIZED_MSG         = "unauthorized"
	MISSING_COOKIE_MSG       = "missing cookie"
	INVALID_COOKIE_MSG       = "invalid cookie"
	INVALID_PAYLOAD_MSG      = "invalid payload"
	INVALID_CONTENT_TYPE_MSG = "invalid content type"
	INVALID_URL_MSG          = "invalid url"
	ENTITY_EXISTS_MSG        = "entity already exists"
)

type Pool struct {
	*pgxpool.Pool
}

type Config struct {
	JWTSecret          string
	BaseURL            string
	GithubClientID     string
	GithubClientSecret string
	GoogleClientID     string
	GoogleClientSecret string
	FrontendURL        string
	GroqModel          string
	GroqAPIKey         string
	GroqBaseURL        string
	FishAudioAPIKey    string
	FishAudioBaseURL   string
	DeepgramAPIKey     string
	DeepgramBaseURL    string
	DeepgramModel      string
	GoogleTokenURL     string
	GithubTokenURL     string
	GoogleUserInfoURL  string
	GithubUserInfoURL  string
	StoragePath        string
	AdminToken         string
}

type HttpErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func HttpError(code int, msg string) []byte {
	return []byte(`{"error": {"code": ` + strconv.Itoa(code) + `, "message": "` + msg + `"}}`)
}

func GetLogger(ctx *gin.Context) *zap.Logger {
	if l, ok := ctx.Get("logger"); ok {
		switch v := l.(type) {
		case *zap.Logger:
			return v
		default:
			panic("logger is not a *zap.Logger")
		}
	} else {
		panic("no logger found in context")
	}
}

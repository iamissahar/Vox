// @title           Vox AI SI API
// @version         1.0
// @description     Vox API server
// @host            api.bogdanantonovich.com
// @BasePath        /vox
// @schemes         https

// @contact.name    Bogdan Antonovich
// @contact.email   programjibogdan@gmail.com

// @securityDefinitions.apikey  CookieAuth
// @in                          cookie
// @name                        access_token
//
// @securityDefinitions.apikey  CookieAuthRefresh
// @in                          cookie
// @name                        refresh_token
//
// @securityDefinitions.apikey AdminAuth
// @in header
// @name Authorization

package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"vox/internal"
	lokisync "vox/pkg/loki"
	"vox/pkg/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	infoLogFile  = "/logs/info.log"
	errorLogFile = "/logs/error.log"
)

func syncLogger(logger *zap.Logger) {
	if err := logger.Sync(); err != nil {
		panic(err)
	}
}

func closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		panic(err)
	}
}

func readFile(filepath string) string {
	body, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(body))
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Errorf("environment variable \"%s\" not set", key))
	}
	return val
}

func newConfig() models.Config {
	return models.Config{
		JWTSecret:          readFile(getEnv("JWT_SECRET_FILE")),
		GithubClientID:     readFile(getEnv("GH_CLIENT_ID_FILE")),
		GithubClientSecret: readFile(getEnv("GH_CLIENT_SECRET_FILE")),
		GoogleClientID:     readFile(getEnv("GOOGLE_CLIENT_ID_FILE")),
		GoogleClientSecret: readFile(getEnv("GOOGLE_CLIENT_SECRET_FILE")),
		GroqAPIKey:         readFile(getEnv("GROQ_API_KEY_FILE")),
		FishAudioAPIKey:    readFile(getEnv("FISH_AUDIO_API_KEY_FILE")),
		DeepgramAPIKey:     readFile(getEnv("DEEPGRAM_API_KEY_FILE")),
		AdminToken:         readFile(getEnv("ADMIN_TOKEN_FILE")),
		BaseURL:            getEnv("BASE_URL"),
		FrontendURL:        getEnv("FRONTEND_URL"),
		GroqModel:          getEnv("GROQ_MODEL"),
		GroqBaseURL:        getEnv("GROQ_BASE_URL"),
		FishAudioBaseURL:   getEnv("FISH_AUDIO_BASE_URL"),
		DeepgramBaseURL:    getEnv("DEEPGRAM_BASE_URL"),
		DeepgramModel:      getEnv("DEEPGRAM_MODEL"),
		GoogleTokenURL:     getEnv("GOOGLE_TOKEN_URL"),
		GithubTokenURL:     getEnv("GH_TOKEN_URL"),
		GoogleUserInfoURL:  getEnv("GOOGLE_USER_INFO_URL"),
		GithubUserInfoURL:  getEnv("GH_USER_INFO_URL"),
		StoragePath:        getEnv("STORAGE_PATH"),
	}
}

func newPool(ctx context.Context, dsn string) *pgxpool.Pool {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(fmt.Errorf("parse pgxpool config: %w", err))
	}

	maxcons, err := strconv.Atoi(getEnv("PGX_MAX_CONNS"))
	if err != nil {
		panic(fmt.Errorf("invalid PGX_MAX_CONNS: %w", err))
	}
	mincons, err := strconv.Atoi(getEnv("PGX_MIN_CONNS"))
	if err != nil {
		panic(fmt.Errorf("invalid PGX_MIN_CONNS: %w", err))
	}

	cfg.MaxConns = int32(maxcons)
	cfg.MinConns = int32(mincons)
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Errorf("create pgxpool: %w", err))
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		panic(fmt.Errorf("ping postgres: %w", err))
	}

	return pool
}

func buildDBURL() string {
	var b strings.Builder
	b.WriteString("postgres://")
	b.WriteString(readFile(getEnv("POSTGRES_USER_FILE")))
	b.WriteString(":")
	b.WriteString(readFile(getEnv("POSTGRES_PASSWORD_FILE")))
	b.WriteString("@")
	b.WriteString(readFile(getEnv("POSTGRES_HOST_FILE")))
	b.WriteString(":")
	b.WriteString(readFile(getEnv("POSTGRES_PORT_FILE")))
	b.WriteString("/")
	b.WriteString(readFile(getEnv("POSTGRES_DB_NAME_FILE")))
	b.WriteString("?sslmode=")
	b.WriteString(readFile(getEnv("POSTGRES_SSLMODE_FILE")))
	return b.String()
}

func newLogger() (*zap.Logger, zap.AtomicLevel, *os.File, *os.File) {
	atom := zap.NewAtomicLevel()
	if getEnv("MODE") == "prod" {
		atom.SetLevel(zapcore.InfoLevel)
	} else {
		atom.SetLevel(zapcore.DebugLevel)
	}

	loggerCfg := zapcore.EncoderConfig{
		TimeKey:      "timestamp",
		LevelKey:     "level",
		MessageKey:   "message",
		CallerKey:    "caller",
		EncodeLevel:  zapcore.CapitalColorLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
	}

	fileCfg := loggerCfg
	fileCfg.EncodeTime = zapcore.RFC3339TimeEncoder

	infoFile, err := os.OpenFile(infoLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		panic(fmt.Errorf("open info.log: %w", err))
	}

	errFile, err := os.OpenFile(errorLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		panic(fmt.Errorf("open error.log: %w", err))
	}

	lokiSyncer := lokisync.NewBuffered(os.Getenv("LOKI_URL"), "vox", 1000)

	core := zapcore.NewTee(
		// stdout — JSON, ISO8601, управляемый уровень
		zapcore.NewCore(
			zapcore.NewJSONEncoder(loggerCfg),
			zapcore.AddSync(os.Stdout),
			atom,
		),
		// Loki — JSON, ISO8601, управляемый уровень
		zapcore.NewCore(
			zapcore.NewJSONEncoder(loggerCfg),
			zapcore.AddSync(lokiSyncer),
			atom,
		),
		// app.log — Console, RFC3339, управляемый уровень
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(fileCfg),
			zapcore.AddSync(infoFile),
			atom,
		),
		// error.log — JSON, RFC3339, только Error+
		zapcore.NewCore(
			zapcore.NewJSONEncoder(fileCfg),
			zapcore.AddSync(errFile),
			zap.NewAtomicLevelAt(zapcore.ErrorLevel),
		),
	)

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), atom, infoFile, errFile
}

func main() {
	logger, atom, infoFile, errFile := newLogger()
	defer syncLogger(logger)
	defer closeFile(infoFile)
	defer closeFile(errFile)

	cfg := newConfig()
	url := buildDBURL()
	fmt.Println(url)
	pool := models.Pool{Pool: newPool(context.Background(), buildDBURL())}
	internal.NewRouter(&cfg, &pool, logger, atom)
}

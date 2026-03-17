package internal

import (
	"net/http"
	"time"
	"vox/internal/admin/logs"
	"vox/internal/auth"
	"vox/internal/hub"
	"vox/internal/user"
	"vox/internal/user/voice"
	"vox/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					zap.Any("error", r),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.Stack("stacktrace"),
				)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func zaplogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		requestLogger := logger.With(
			zap.String("request_id", uuid.New().String()),
			zap.String("ip", c.ClientIP()),
		)
		c.Set("logger", requestLogger)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		if query != "" {
			path = path + "?" + query
		}

		requestLogger.Info("incoming request",
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}

func NewRouter(cfg *models.Config, pool *models.Pool, logger *zap.Logger, atom zap.AtomicLevel) {
	engine := gin.New()

	engine.Use(zaplogger(logger))
	engine.Use(recovery(logger))

	userAPI := user.UserAPI{DB: user.NewUserDB(pool)}
	authAPI := auth.AuthAPI{DB: auth.NewAuthDB(pool), Cfg: cfg}
	hubAPI := hub.HubAPI{DB: hub.NewHubDB(pool), Cfg: cfg}
	voiceAPI := voice.VoiceAPI{DB: voice.NewVoiceDB(pool), Cfg: cfg}
	logsAPI := logs.LogsAPI{Atomic: atom}

	// public routes
	authGroup := engine.Group("/auth")
	{
		providerGroup := authGroup.Group("/:provider")
		{
			providerGroup.GET("/callback", authAPI.ProviderCallbackHandler)
			providerGroup.GET("/login", authAPI.LoginViaProviderHandler)
		}
		authGroup.POST("/login", authAPI.LoginHandler)
		authGroup.POST("/sign_up", authAPI.SignUpHandler)
	}

	hubGroup := engine.Group("/hub/:hub_id")
	hubGroup.Use(hubAPI.IsHubIDValid)
	{
		privateHubGroup := hubGroup.Group("/")
		privateHubGroup.Use(authAPI.IsAuthorized)
		privateHubGroup.Use(hubAPI.IsContentTypeValid)
		privateHubGroup.Use(hubAPI.FishSDK)
		{
			privateHubGroup.POST("/publish", hubAPI.PublishHandler)
			privateHubGroup.POST("/new", hubAPI.NewHubHandler)
		}
		hubGroup.GET("/listen", hubAPI.ListenHandler)
	}

	userGroup := engine.Group("/user")
	userGroup.Use(authAPI.IsAuthorized)
	{
		userGroup.GET("/info", userAPI.InfoHandler)
		voiceGroup := userGroup.Group("/voice")
		{
			voiceGroup.POST("/new", voiceAPI.ReferenceHandler)
		}
	}

	// helpers and admin routes
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	adminGroup := engine.Group("/")
	adminGroup.Use(authAPI.IsAdmin)
	{
		adminGroup.PUT("/admin/logs/level", logsAPI.LevelHandler)
	}

	logger.Info("API started", zap.String("env", "prod"), zap.Int("port", 9081))
	_ = engine.Run(":9081")
}

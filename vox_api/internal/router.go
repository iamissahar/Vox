package internal

import (
	"context"
	"net/http"
	"time"
	"vox/internal/admin/logs"
	"vox/internal/auth"
	"vox/internal/hub"
	"vox/internal/user"
	"vox/internal/user/voice"
	"vox/pkg/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	limiter "github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
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

// healthHandler godoc
// @Summary      Health check
// @Tags         service
// @Success      200
// @Router       /health [get]
func healthHandler(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}

func timeout(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}

// TODO: migrate rate limiter store to Redis when horizontal scaling is needed
// TODO: add separate strict rate limit (5-M) for /auth/login and /auth/sign_up
// TODO: add API versioning (/v1/)
// TODO: split hub getOrCreate into two routes: POST /hub (create) and GET /hub/:hub_id (reconnect)
// TODO: add heartbeat mechanism for /publish and /listen to detect dead connections
// TODO: move hardcoded port to cfg
// TODO: handle engine.Run() error instead of discarding it
// TODO: handle limiter.NewRateFromFormatted() error instead of discarding it
func NewRouter(cfg *models.Config, pool *models.Pool, logger *zap.Logger, atom zap.AtomicLevel) {
	userAPI := user.UserAPI{DB: user.NewUserDB(pool)}
	authAPI := auth.AuthAPI{DB: auth.NewAuthDB(pool), Cfg: cfg}
	hubAPI := hub.HubAPI{DB: hub.NewHubDB(pool), Cfg: cfg}
	voiceAPI := voice.VoiceAPI{DB: voice.NewVoiceDB(pool), Cfg: cfg}
	logsAPI := logs.LogsAPI{Atomic: atom}

	engine := gin.New()

	engine.Use(zaplogger(logger))
	engine.Use(recovery(logger))
	rate, _ := limiter.NewRateFromFormatted("100-M")
	store := memory.NewStore()
	engine.Use(mgin.NewMiddleware(limiter.New(store, rate)))
	engine.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	})

	corsGroups := engine.Group("/")
	corsGroups.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"https://bogdanantonovich.com",
			"https://www.bogdanantonovich.com",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           86400 * time.Second,
	}))

	authGroup := corsGroups.Group("/auth")
	authGroup.Use(timeout)
	{
		providerGroup := authGroup.Group("/:provider")
		{
			providerGroup.GET("/callback", authAPI.ProviderCallbackHandler)
			providerGroup.GET("/login", authAPI.LoginViaProviderHandler)
		}
		authGroup.POST("/login", authAPI.LoginHandler)
		authGroup.POST("/sign_up", authAPI.SignUpHandler)
	}

	hubGroup := corsGroups.Group("/hub/:hub_id")
	hubGroup.Use(hubAPI.IsHubIDValid)
	{
		privateHubGroup := hubGroup.Group("/")
		privateHubGroup.Use(authAPI.IsAuthorized)
		privateHubGroup.Use(hubAPI.IsContentTypeValid)
		privateHubGroup.Use(hubAPI.FishSDK)
		{
			privateHubGroup.POST("/publish", hubAPI.PublishHandler)
			privateHubGroup2 := privateHubGroup.Group("/")
			privateHubGroup2.Use(timeout)
			{
				privateHubGroup2.POST("/new", hubAPI.NewHubHandler)
			}
		}
		hubGroup.GET("/listen", hubAPI.ListenHandler)
	}

	userGroup := corsGroups.Group("/user")
	userGroup.Use(authAPI.IsAuthorized)
	{
		userGroup.GET("/info", timeout, userAPI.InfoHandler)
		voiceGroup := userGroup.Group("/voice")
		{
			voiceGroup.POST("/new", voiceAPI.ReferenceHandler)
		}
	}

	adminGroup := corsGroups.Group("/admin")
	adminGroup.Use(authAPI.IsAdmin)
	{
		adminGroup.PUT("/logs/level", logsAPI.LevelHandler)
	}

	// public helpers
	helpersGroup := engine.Group("/")
	helpersGroup.Use(timeout)
	{
		helpersGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		helpersGroup.GET("/health", healthHandler)
	}

	logger.Info("API started", zap.String("env", "prod"), zap.Int("port", 9081))
	_ = engine.Run(":9081")
}

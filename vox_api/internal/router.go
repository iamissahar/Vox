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

func NewRouter(cfg *models.Config, pool *models.Pool, logger *zap.Logger, atom zap.AtomicLevel) {
	userAPI := user.UserAPI{DB: user.NewUserDB(pool)}
	authAPI := auth.AuthAPI{DB: auth.NewAuthDB(pool), Cfg: cfg}
	hubAPI := hub.HubAPI{DB: hub.NewHubDB(pool), Cfg: cfg, MGR: hub.NewManager()}
	voiceAPI := voice.VoiceAPI{DB: voice.NewVoiceDB(pool), Cfg: cfg}
	logsAPI := logs.LogsAPI{Atomic: atom}
	hostAndHubs := hub.NewHostAndHubs()

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

	corsGroups.GET("/auth/:provider/callback", timeout, authAPI.ProviderCallbackHandler)
	corsGroups.GET("/auth/:provider/login", timeout, authAPI.LoginViaProviderHandler)
	corsGroups.POST("/auth/login", timeout, authAPI.LoginHandler)
	corsGroups.POST("/auth/sign_up", timeout, authAPI.SignUpHandler)

	corsGroups.POST("/hub", timeout, authAPI.IsAuthorized, hubAPI.PutCache(hostAndHubs), hubAPI.NewHubHandler)
	corsGroups.DELETE("/hub/:hub_id", timeout, authAPI.IsAuthorized, hubAPI.IsHubIDValid, hubAPI.PutCache(hostAndHubs), hubAPI.DeleteHubHandler)
	corsGroups.GET("/hub/:hub_id/listen", hubAPI.IsHubIDValid, hubAPI.ListenHandler)
	corsGroups.GET("/hub/:hub_id/reconnect", timeout, authAPI.IsAuthorized, hubAPI.IsHubIDValid, hubAPI.ReconnectHandler)
	corsGroups.POST("/hub/:hub_id/publish", authAPI.IsAuthorized, hubAPI.IsHubIDValid, hubAPI.IsContentTypeValid, hubAPI.FishSDK, hubAPI.PublishHandler)

	corsGroups.GET("/user/info", timeout, authAPI.IsAuthorized, userAPI.InfoHandler)
	corsGroups.GET("/user/hubs", timeout, authAPI.IsAuthorized, hubAPI.PutCache(hostAndHubs), userAPI.HubsHandler)
	corsGroups.POST("/user/voice", authAPI.IsAuthorized, hubAPI.IsContentTypeValid, voiceAPI.ReferenceHandler)

	corsGroups.PUT("/admin/logs/level", timeout, authAPI.IsAdmin, logsAPI.LevelHandler)

	engine.GET("/swagger/*any", timeout, ginSwagger.WrapHandler(swaggerFiles.Handler))
	engine.GET("/health", timeout, healthHandler)

	logger.Info("API started", zap.String("env", "prod"), zap.Int("port", 9081))
	_ = engine.Run(":9081")
}

// func NewRouter(cfg *models.Config, pool *models.Pool, logger *zap.Logger, atom zap.AtomicLevel) {
// 	userAPI := user.UserAPI{DB: user.NewUserDB(pool)}
// 	authAPI := auth.AuthAPI{DB: auth.NewAuthDB(pool), Cfg: cfg}
// 	hubAPI := hub.HubAPI{DB: hub.NewHubDB(pool), Cfg: cfg}
// 	voiceAPI := voice.VoiceAPI{DB: voice.NewVoiceDB(pool), Cfg: cfg}
// 	logsAPI := logs.LogsAPI{Atomic: atom}
// 	hostAndHubs := hub.NewHostAndHubs()

// 	engine := gin.New()
// 	engine.RedirectTrailingSlash = false

// 	engine.Use(zaplogger(logger))
// 	engine.Use(recovery(logger))
// 	rate, _ := limiter.NewRateFromFormatted("100-M")
// 	store := memory.NewStore()
// 	engine.Use(mgin.NewMiddleware(limiter.New(store, rate)))
// 	engine.Use(func(c *gin.Context) {
// 		c.Header("X-Content-Type-Options", "nosniff")
// 		c.Header("X-Frame-Options", "DENY")
// 		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
// 		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
// 		c.Next()
// 	})

// 	corsGroups := engine.Group("/")
// 	corsGroups.Use(cors.New(cors.Config{
// 		AllowOrigins: []string{
// 			"https://bogdanantonovich.com",
// 			"https://www.bogdanantonovich.com",
// 		},
// 		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
// 		AllowHeaders:     []string{"Content-Type", "Authorization"},
// 		AllowCredentials: true,
// 		MaxAge:           86400 * time.Second,
// 	}))

// 	authGroup := corsGroups.Group("/auth")
// 	authGroup.Use(timeout)
// 	{
// 		providerGroup := authGroup.Group("/:provider")
// 		{
// 			providerGroup.GET("/callback", authAPI.ProviderCallbackHandler)
// 			providerGroup.GET("/login", authAPI.LoginViaProviderHandler)
// 		}
// 		authGroup.POST("/login", authAPI.LoginHandler)
// 		authGroup.POST("/sign_up", authAPI.SignUpHandler)
// 	}

// 	hubGroup := corsGroups.Group("/hub")
// 	{
// 		withTimeOutGroup := hubGroup.Group("/")
// 		withTimeOutGroup.Use(timeout)
// 		withTimeOutGroup.Use(authAPI.IsAuthorized)
// 		// change it for something else in the future
// 		withTimeOutGroup.Use(hubAPI.PutCache(hostAndHubs))
// 		{
// 			withTimeOutGroup.POST("", hubAPI.NewHubHandler)
// 		}

// 		withHubIDGroup := hubGroup.Group("/:hub_id")
// 		withHubIDGroup.Use(hubAPI.IsHubIDValid)
// 		{
// 			withHubIDGroup.DELETE("", timeout, hubAPI.PutCache(hostAndHubs), hubAPI.DeleteHubHandler)
// 			withHubIDGroup.GET("/listen", hubAPI.ListenHandler)

// 			privateHubGroup := withHubIDGroup.Group("/")
// 			privateHubGroup.Use(authAPI.IsAuthorized)
// 			privateHubGroup.Use(hubAPI.IsContentTypeValid)
// 			privateHubGroup.Use(hubAPI.FishSDK)
// 			{
// 				privateWithTimeoutGroup := privateHubGroup.Group("/")
// 				privateWithTimeoutGroup.Use(timeout)
// 				{
// 					privateWithTimeoutGroup.GET("/reconnect", hubAPI.ReconnectHandler)
// 				}

// 				privateHubGroup.POST("/publish", hubAPI.PublishHandler)
// 			}
// 		}

// 	}

// 	userGroup := corsGroups.Group("/user")
// 	userGroup.Use(authAPI.IsAuthorized)
// 	{
// 		userWithTimeoutGroup := userGroup.Group("/")
// 		userWithTimeoutGroup.Use(timeout)
// 		{
// 			userWithTimeoutGroup.GET("/info", userAPI.InfoHandler)
// 			userWithTimeoutGroup.GET("/hubs", hubAPI.PutCache(hostAndHubs), userAPI.HubsHandler)
// 		}
// 		voiceGroup := userGroup.Group("/voice")
// 		{
// 			voiceGroup.POST("/new", voiceAPI.ReferenceHandler)
// 		}
// 	}

// 	adminGroup := corsGroups.Group("/admin")
// 	adminGroup.Use(authAPI.IsAdmin)
// 	{
// 		adminGroup.PUT("/logs/level", logsAPI.LevelHandler)
// 	}

// 	// public helpers
// 	helpersGroup := engine.Group("/")
// 	helpersGroup.Use(timeout)
// 	{
// 		helpersGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
// 		helpersGroup.GET("/health", healthHandler)
// 	}

// 	logger.Info("API started", zap.String("env", "prod"), zap.Int("port", 9081))
// 	_ = engine.Run(":9081")
// }

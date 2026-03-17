func NewRouter(cfg *models.Config, pool *models.Pool, logger *zap.Logger, atom zap.AtomicLevel) {
	engine := gin.New()
	engine.Use(zaplogger(logger))
	engine.Use(recovery(logger))

	// ── CORS ──────────────────────────────────────────────────────────────────
	// Allowed origins come from config so you can set them per environment
	// via env var / config file without redeploying.
	// Example: cfg.AllowedOrigins = []string{"https://yourdomain.com"}
	engine.Use(middleware.CORS(cfg.AllowedOrigins))

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

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	adminGroup := engine.Group("/")
	adminGroup.Use(authAPI.IsAdmin)
	{
		adminGroup.PUT("/admin/logs/level", logsAPI.LevelHandler)
	}

	logger.Info("API started", zap.String("env", "prod"), zap.Int("port", 9081))
	_ = engine.Run(":9081")
}

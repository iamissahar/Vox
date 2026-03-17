// В твоём models.Config добавь одно поле:

type Config struct {
	// ... существующие поля ...

	AllowedOrigins []string `mapstructure:"ALLOWED_ORIGINS"`
}

// Пример в .env / docker secret:
//   ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
//
// Если используешь Viper, он автоматически распарсит через запятую в []string.
// Если читаешь вручную через os.Getenv:

// cfg.AllowedOrigins = strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")

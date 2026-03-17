package main

func main() {
	// var encoderCfg = zapcore.EncoderConfig{
	// 	TimeKey:      "timestamp",
	// 	LevelKey:     "level",
	// 	MessageKey:   "message",
	// 	CallerKey:    "caller",
	// 	EncodeTime:   zapcore.RFC3339TimeEncoder,
	// 	EncodeLevel:  zapcore.CapitalColorLevelEncoder,
	// 	EncodeCaller: zapcore.ShortCallerEncoder,
	// }
	// var core = zapcore.NewTee(
	// 	zapcore.NewCore(
	// 		zapcore.NewConsoleEncoder(encoderCfg),
	// 		zapcore.AddSync(os.Stdout),
	// 		zapcore.DebugLevel,
	// 	),
	// )

	// logger := zap.New(core)
	// defer func() {
	// 	if err := logger.Sync(); err != nil {
	// 		panic(err)
	// 	}
	// }()

	// var cfg = models.Config{}

	// var pool = models.Pool{}

	// internal.NewRouter(&cfg, &pool, logger)
}

package logging

import (
	"time"

	"github.com/rs/zerolog"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// CreateDebugHTTP to initialize the debug HTTP logger
func CreateDebugHTTP(filename string, cfg LumberjackConfig) (*zerolog.Logger, error) {
	zerolog.TimeFieldFormat = time.RFC3339
	z := zerolog.New(&lumberjack.Logger{
		Filename:   filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	})
	logger := z.With().Caller().Timestamp().Logger()
	return &logger, nil
}

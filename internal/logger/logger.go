package logger

import (
	"sync"

	"go.uber.org/zap"
)

var (
	once   sync.Once
	logger *zap.Logger
)

func InitLogger() {
	once.Do(func() {
		logger, _ = zap.NewProduction()
	})
}

func Info(msg string, fields ...zap.Field) {
	if logger == nil {
		return
	}

	logger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	if logger == nil {
		return
	}

	logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	if logger == nil {
		return
	}

	logger.Fatal(msg, fields...)
}

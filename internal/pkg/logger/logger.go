package logger

import (
	"context"

	"go.uber.org/zap"
)

type loggerKey string

const key loggerKey = "contextLogger"

// ToContext помещает logger в контекст
func ToContext(ctx context.Context, sugarLogger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, key, sugarLogger)
}

// Infof ...
func Infof(ctx context.Context, format string, args ...any) {
	logger := FromContext(ctx)
	logger.Infof(format, args...)
}

// Errorf ...
func Errorf(ctx context.Context, format string, args ...any) {
	logger := FromContext(ctx)
	logger.Errorf(format, args...)
}

var defaultSugarLogger, _ = zap.NewDevelopment()

// FromContext извлекает logger из контекста
func FromContext(ctx context.Context) *zap.SugaredLogger {
	logger := ctx.Value(key)
	sugarLogger, ok := logger.(*zap.SugaredLogger)
	if !ok {
		return defaultSugarLogger.Sugar()
	}

	return sugarLogger
}

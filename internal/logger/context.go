package logger

import (
	"context"
	"log/slog"
)

type ctxLogger struct{}

func ContextWithLogger(ctx context.Context, sLogger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxLogger{}, sLogger)
}

func loggerFromContext(ctx context.Context) *slog.Logger {
	if sLogger, ok := ctx.Value(ctxLogger{}).(*slog.Logger); ok {
		return sLogger
	}
	return slog.Default()
}

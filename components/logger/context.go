package logger

import (
	"context"

	"github.com/rs/zerolog"
)

type loggerKey struct{}

func UpdateContext(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func FromContext(ctx context.Context) zerolog.Logger {
	lg, ok := ctx.Value(loggerKey{}).(zerolog.Logger)
	if !ok {
		return DefaultLogger()
	}
	return lg
}

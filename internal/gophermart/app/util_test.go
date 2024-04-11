package app_test

import (
	"context"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/google/uuid"
)

var _ domain.Logger = (*testLogger)(nil)

type testLogger struct {
}

func (testLogger) Infow(msg string, keysAndValues ...any)  {}
func (testLogger) Errorw(msg string, keysAndValues ...any) {}

func EnrichTestContext(ctx context.Context) context.Context {
	requestUUID := uuid.New()
	return domain.EnrichWithRequestIDLogger(ctx, requestUUID, &testLogger{})
}

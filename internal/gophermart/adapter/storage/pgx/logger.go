package pgx

import (
	"context"
	"log"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/jackc/pgx/v5/tracelog"
)

type loggerAdapter struct {
	logger domain.Logger
}

func (la *loggerAdapter) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {

	keyAndValues := make([]any, 2*len(data))
	i := 0
	for k, v := range data {
		keyAndValues[i] = k
		keyAndValues[i+1] = v
		i += 2
	}

	switch level {
	case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
		la.logger.Debugw(msg, keyAndValues...)
	case tracelog.LogLevelInfo, tracelog.LogLevelWarn:
		la.logger.Infow(msg, keyAndValues...)
	case tracelog.LogLevelError:
		la.logger.Errorw(msg, keyAndValues...)
	default:
		log.Printf("invalid level %d", la)
	}
}

package domain

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ContextKey string

const KeyLogger = ContextKey("Logger")
const KeyAuthData = ContextKey("AuthData")

func NewContext(ctx context.Context, authData *AuthData, requestID uuid.UUID, zL *zap.SugaredLogger) context.Context {
	loggerCtx := context.WithValue(ctx, KeyLogger, newLogger(authData.UserID, requestID, zL))
	authDataCtx := context.WithValue(loggerCtx, KeyAuthData, authData)
	return authDataCtx
}

func GetAuthData(ctx context.Context) (*AuthData, error) {
	if v := ctx.Value(KeyAuthData); v != nil {
		authData, ok := v.(*AuthData)
		if !ok {
			return nil, fmt.Errorf("%w: unexpected authData type", ErrUserIsNotAuthorized)
		}
		return authData, nil
	}
	return nil, fmt.Errorf("%w: can't extract authData", ErrUserIsNotAuthorized)
}

func GetLogger(ctx context.Context) (Logger, error) {
	if v := ctx.Value(KeyLogger); v != nil {
		lg, ok := v.(Logger)
		if !ok {
			return nil, fmt.Errorf("%w: unexpected logger type", ErrServerInternal)
		}
		return lg, nil
	}
	return nil, fmt.Errorf("%w: can't extract logger", ErrServerInternal)
}

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . Logger
type Logger interface {
	Infow(msg string, keysAndValues ...any)
	Errorw(msg string, err error, keysAndValues ...any)
}

func newLogger(userID int, requestID uuid.UUID, zL *zap.SugaredLogger) Logger {
	return &logger{
		UserID:        userID,
		RequestID:     requestID,
		SugaredLogger: zL,
	}
}

type logger struct {
	UserID    int
	RequestID uuid.UUID
	*zap.SugaredLogger
}

func (l *logger) Infow(msg string, keysAndValues ...any) {
	keysAndValues = append(keysAndValues, "userID", strconv.Itoa(l.UserID), "requestID", l.RequestID.String())
	l.SugaredLogger.Infow(msg, keysAndValues...)
}

func (l *logger) Errorw(msg string, err error, keysAndValues ...any) {
	keysAndValues = append(keysAndValues, "error", err.Error(), "userID", strconv.Itoa(l.UserID), "requestID", l.RequestID.String())
	l.SugaredLogger.Infow(msg, keysAndValues...)
}

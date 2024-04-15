package domain

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"

	_ "github.com/golang/mock/gomock"        // обязательно, требуется в сгенерированных mock-файлах,
	_ "github.com/golang/mock/mockgen/model" // обязательно для корректного запуска mockgen
)

type ContextKey string

const KeyRequestID = ContextKey("RequestID")
const KeyLogger = ContextKey("Logger")
const KeyAuthData = ContextKey("AuthData")

const LoggerKeyRequestID = "requestID"
const LoggerKeyUserID = "userID"

//go:generate mockgen -destination "./mocks/$GOFILE" -package mocks . Logger
type Logger interface {
	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
}

func EnrichWithRequestIDLogger(ctx context.Context, requestID uuid.UUID, logger Logger) context.Context {
	requestIDLogger := &requestIDLogger{
		internalLogger: logger,
		requestID:      requestID.String(),
	}
	resultCtx := context.WithValue(ctx, KeyLogger, requestIDLogger)
	return resultCtx
}

func EnrichWithAuthData(ctx context.Context, authData *AuthData) (context.Context, error) {

	if authData == nil {
		return ctx, fmt.Errorf("%w: authData is nil", ErrServerInternal)
	}

	curLogger, err := GetLogger(ctx)
	if err != nil {
		return ctx, fmt.Errorf("can't get logger %w", err)
	}

	resultCtx := context.WithValue(ctx, KeyAuthData, authData)
	resultCtx = context.WithValue(resultCtx, KeyLogger, &userIDLogger{
		internalLogger: curLogger,
		userID:         strconv.Itoa(authData.UserID),
	})
	return resultCtx, nil
}

func GetRequestID(ctx context.Context) (uuid.UUID, error) {
	if v := ctx.Value(KeyRequestID); v != nil {
		requestID, ok := v.(uuid.UUID)
		if !ok {
			return uuid.Nil, fmt.Errorf("%w: unexpected requestID type", ErrServerInternal)
		}
		return requestID, nil
	}
	return uuid.Nil, fmt.Errorf("%w: can't extract requestID", ErrServerInternal)
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

func GetUserID(ctx context.Context) (int, error) {
	a, err := GetAuthData(ctx)
	if err != nil {
		return -1, fmt.Errorf("%w: can't get userID", err)
	}
	return a.UserID, nil
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

var _ Logger = (*requestIDLogger)(nil)

type requestIDLogger struct {
	requestID      string
	internalLogger Logger
}

func (l *requestIDLogger) Debugw(msg string, keysAndValues ...any) {
	keysAndValues = append(keysAndValues, LoggerKeyRequestID, l.requestID)
	l.internalLogger.Debugw(msg, keysAndValues...)
}

func (l *requestIDLogger) Infow(msg string, keysAndValues ...any) {
	keysAndValues = append(keysAndValues, LoggerKeyRequestID, l.requestID)
	l.internalLogger.Infow(msg, keysAndValues...)
}

func (l *requestIDLogger) Errorw(msg string, keysAndValues ...any) {
	keysAndValues = append(keysAndValues, LoggerKeyRequestID, l.requestID)
	l.internalLogger.Infow(msg, keysAndValues...)
}

var _ Logger = (*userIDLogger)(nil)

type userIDLogger struct {
	userID         string
	internalLogger Logger
}

func (l *userIDLogger) Infow(msg string, keysAndValues ...any) {
	keysAndValues = append(keysAndValues, LoggerKeyUserID, l.userID)
	l.internalLogger.Infow(msg, keysAndValues...)
}

func (l *userIDLogger) Debugw(msg string, keysAndValues ...any) {
	keysAndValues = append(keysAndValues, LoggerKeyUserID, l.userID)
	l.internalLogger.Infow(msg, keysAndValues...)
}

func (l *userIDLogger) Errorw(msg string, keysAndValues ...any) {
	keysAndValues = append(keysAndValues, LoggerKeyUserID, l.userID)
	l.internalLogger.Infow(msg, keysAndValues...)
}

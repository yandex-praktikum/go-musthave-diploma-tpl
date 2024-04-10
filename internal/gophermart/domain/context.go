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

const KeyLogger = ContextKey("Logger")
const KeyAuthData = ContextKey("AuthData")

const LoggerKeyRequestID = "requestID"
const LoggerKeyUserID = "userID"

//go:generate mockgen -destination "./mocks/$GOFILE" -package mocks . Logger
type Logger interface {
	Infow(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
}

func EnrichContext(ctx context.Context, authData *AuthData, requestID uuid.UUID, logger Logger) context.Context {
	var userId int
	if authData != nil {
		userId = authData.UserID
	}
	loggerCtx := context.WithValue(ctx, KeyLogger, newLogger(userId, requestID, logger))

	resultCtx := loggerCtx
	if authData != nil {
		resultCtx = context.WithValue(loggerCtx, KeyAuthData, authData)
	}
	return resultCtx
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

var _ Logger = (*logger)(nil)

func newLogger(userID int, requestID uuid.UUID, intLogger Logger) *logger {
	return &logger{
		userID:         userID,
		requestID:      requestID,
		internalLogger: intLogger,
	}
}

type logger struct {
	userID         int
	requestID      uuid.UUID
	internalLogger Logger
}

func (l *logger) Infow(msg string, keysAndValues ...any) {
	if l.userID > 0 {
		keysAndValues = append(keysAndValues, LoggerKeyUserID, strconv.Itoa(l.userID), LoggerKeyRequestID, l.requestID.String())
	} else {
		keysAndValues = append(keysAndValues, LoggerKeyRequestID, l.requestID.String())
	}

	l.internalLogger.Infow(msg, keysAndValues...)
}

func (l *logger) Errorw(msg string, keysAndValues ...any) {
	if l.userID > 0 {
		keysAndValues = append(keysAndValues, LoggerKeyUserID, strconv.Itoa(l.userID), LoggerKeyRequestID, l.requestID.String())
	} else {
		keysAndValues = append(keysAndValues, LoggerKeyRequestID, l.requestID.String())
	}
	l.internalLogger.Infow(msg, keysAndValues...)
}

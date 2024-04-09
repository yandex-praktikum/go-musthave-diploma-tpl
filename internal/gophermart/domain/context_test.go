package domain_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/StasMerzlyakov/go-musthave-diploma-tpl/internal/gophermart/domain"
	"github.com/StasMerzlyakov/go-musthave-diploma-tpl/internal/gophermart/domain/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEnrichContext(t *testing.T) {

	authData := &domain.AuthData{
		UserID: 123,
	}

	userID := strconv.Itoa(authData.UserID)

	requestUUID := uuid.New()
	reqStr := requestUUID.String()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockLogger(ctrl)

	testLoggerFn := func(msg string, keysAndValues ...any) {
		// Проверяем что что при вызове метода логирования добавляется информация о пользователе и requstId
		userIDIsChecked := false
		requestIDIsChecked := false

		for id, v := range keysAndValues {
			switch v := v.(type) {
			case string:
				if v == "userID" {
					require.True(t, id+1 < len(keysAndValues), "userID is not set")
					k := keysAndValues[id+1]
					id, ok := k.(string)
					require.True(t, ok, "userID is not string")
					require.Equal(t, userID, id, "unexpecred userID value")
					userIDIsChecked = true
				}
				if v == "requestID" {
					require.True(t, id+1 < len(keysAndValues), "requestID is not set")
					k := keysAndValues[id+1]
					id, ok := k.(string)
					require.True(t, ok, "requestID is not string")
					require.Equal(t, reqStr, id, "unexpecred requestID value")
					requestIDIsChecked = true
				}
			}
		}
		require.Truef(t, userIDIsChecked && requestIDIsChecked, "userID or requestID is not specified")
	}

	m.EXPECT().Infow(gomock.Any(), gomock.Any()).DoAndReturn(testLoggerFn).AnyTimes()

	m.EXPECT().Errorw(gomock.Any(), gomock.Any()).DoAndReturn(testLoggerFn).AnyTimes()

	ctx := context.Background()

	enrichedCtx := domain.EnrichContext(ctx, authData, requestUUID, m)

	aData, err := domain.GetAuthData(enrichedCtx)

	require.NoError(t, err)

	require.NotNil(t, aData)

	require.Equal(t, authData.UserID, aData.UserID)

	log, err := domain.GetLogger(enrichedCtx)
	require.NoError(t, err)

	log.Errorw("test errorw", "msg", "hello")
	log.Infow("test errorw", "msg", "hello")
}

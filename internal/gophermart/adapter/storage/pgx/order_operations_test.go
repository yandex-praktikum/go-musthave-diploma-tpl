package pgx_test

import (
	"context"
	"testing"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/config"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/storage/pgx"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/stretchr/testify/require"
)

func TestOrderFunctions(t *testing.T) {
	ctx, cancelFN := context.WithCancel(context.Background())

	defer cancelFN()

	connString, err := postgresContainer.ConnectionString(ctx)

	require.NoError(t, err)

	logger := createLogger()
	storage := pgx.NewStorage(ctx, logger, &config.GophermartConfig{
		MaxConns:             5,
		DatabaseUri:          connString,
		ProcessingLimit:      5,
		ProcessingScoreDelta: 10 * time.Second,
	})

	err = storage.Ping(ctx)
	require.NoError(t, err)

	user1 := "user1"

	ldata1 := &domain.LoginData{
		Login: user1,
		Hash:  "",
		Salt:  "",
	}

	user1ID, err := storage.RegisterUser(ctx, ldata1)
	require.NoError(t, err)

	user2 := "user2"

	ldata2 := &domain.LoginData{
		Login: user2,
		Hash:  "",
		Salt:  "",
	}

	user2ID, err := storage.RegisterUser(ctx, ldata2)
	require.NoError(t, err)
	require.NotEqual(t, user1ID, user2ID)

	orderNum := "123456"

	orderData := &domain.OrderData{
		UserID:     user1ID,
		Number:     domain.OrderNumber(orderNum),
		Status:     domain.OrderStratusNew,
		UploadedAt: domain.RFC3339Time(time.Now()),
	}
	err = storage.Upload(ctx, orderData)
	require.NoError(t, err)

	err = storage.Upload(ctx, orderData)
	require.ErrorIs(t, err, domain.ErrOrderNumberAlreadyUploaded)

	orderData.UserID = user2ID
	err = storage.Upload(ctx, orderData)
	require.ErrorIs(t, err, domain.ErrDublicateOrderNumber)

	data, err := storage.ForProcessing(ctx, []domain.OrderStatus{domain.OrderStratusProcessed})
	require.NoError(t, err)
	require.Empty(t, data)

	orderData.Number = "234567"
	err = storage.Upload(ctx, orderData)
	require.NoError(t, err)

	data, err = storage.ForProcessing(ctx, []domain.OrderStatus{domain.OrderStratusNew})
	require.NoError(t, err)

	require.Equal(t, 2, len(data))

	data, err = storage.ForProcessing(ctx, []domain.OrderStatus{domain.OrderStratusNew})
	require.NoError(t, err)

	require.Empty(t, data)

}

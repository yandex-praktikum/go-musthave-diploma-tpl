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

func TestBalanceOperations(t *testing.T) {
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

	defer func() {
		err = clear(ctx)
		require.NoError(t, err)
	}()

	err = storage.Ping(ctx)
	require.NoError(t, err)

	err = clear(ctx)
	require.NoError(t, err)

	login := "user123"
	passHash := "hash"
	salt := "salt"
	ldata, err := storage.GetUserData(ctx, login)
	require.NoError(t, err)
	require.Nil(t, ldata)

	ldata = &domain.LoginData{
		Login: login,
		Hash:  passHash,
		Salt:  salt,
	}

	userID, err := storage.RegisterUser(ctx, ldata)
	require.NoError(t, err)

	bal, err := storage.Balance(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, bal)

	require.Equal(t, 0., bal.Current)

	// Добавим начисление

	orderNum := domain.OrderNumber("1234562")

	now := domain.RFC3339Time(time.Now())

	orderData := &domain.OrderData{
		UserID:     userID,
		Number:     orderNum,
		Status:     domain.OrderStratusNew,
		UploadedAt: now,
	}
	err = storage.Upload(ctx, orderData)
	require.NoError(t, err)

	accrual := domain.Float64Ptr(50.)
	err = storage.UpdateOrder(ctx, orderNum, domain.OrderStratusProcessing, accrual)
	require.NoError(t, err)

	orderData.Accrual = accrual
	orderData.Status = domain.OrderStratusProcessed

	bal.Current = *accrual
	err = storage.UpdateBalanceByOrder(ctx, bal, orderData)
	require.NoError(t, err)

	bal2, err := storage.Balance(ctx, userID)
	require.NoError(t, err)

	require.Equal(t, bal.Current, bal2.Current)

	err = storage.UpdateBalanceByOrder(ctx, bal, orderData)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

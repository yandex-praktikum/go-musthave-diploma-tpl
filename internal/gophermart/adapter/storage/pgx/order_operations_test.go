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

	defer func() {
		err = clear(ctx)
		require.NoError(t, err)
	}()

	err = storage.Ping(ctx)
	require.NoError(t, err)

	err = clear(ctx)
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

	orderNum := domain.OrderNumber("1234561")

	now := domain.RFC3339Time(time.Now())

	orderData := &domain.OrderData{
		UserID:     user1ID,
		Number:     orderNum,
		Status:     domain.OrderStratusNew,
		UploadedAt: now,
	}
	err = storage.Upload(ctx, orderData)
	require.NoError(t, err)

	orderDatas, err := storage.Orders(ctx, user1ID)
	require.NoError(t, err)

	require.Equal(t, 1, len(orderDatas))

	oD := orderDatas[0]
	require.Equal(t, user1ID, oD.UserID)
	require.Equal(t, orderNum, oD.Number)
	require.Equal(t, domain.OrderStratusNew, oD.Status)

	require.Nil(t, oD.Accrual)
	// при сохранении в PostgreSQL теряется точность до миллисекунды
	//  2024-04-17 14:36:19.18034987 +0000 UTC - до сохраенния
	//  2024-04-17 14:36:19.180349 +0000 +0000 - после
	// require.True(t, time.Time(now).Equal(time.Time(oD.UploadedAt)))

	accrualVal := 10.9
	err = storage.UpdateOrder(ctx, orderNum, domain.OrderStratusProcessing, &accrualVal)
	require.NoError(t, err)

	orderDatas, err = storage.Orders(ctx, user1ID)
	require.NoError(t, err)

	require.Equal(t, 1, len(orderDatas))

	oD = orderDatas[0]
	require.Equal(t, user1ID, oD.UserID)
	require.Equal(t, orderNum, oD.Number)
	require.Equal(t, domain.OrderStratusProcessing, oD.Status)
	require.NotNil(t, oD.Accrual)
	require.Equal(t, accrualVal, *oD.Accrual)

	orderDatas, err = storage.Orders(ctx, user2ID)
	require.NoError(t, err)
	require.Empty(t, orderDatas)

	// Проверка уникальности номера
	err = storage.Upload(ctx, orderData)
	require.ErrorIs(t, err, domain.ErrOrderNumberAlreadyUploaded)

	orderData.UserID = user2ID
	err = storage.Upload(ctx, orderData)
	require.ErrorIs(t, err, domain.ErrDublicateOrderNumber)

	orderDatas, err = storage.Orders(ctx, user2ID)
	require.NoError(t, err)
	require.Empty(t, orderDatas)

	// Проверка получения данных по статусу
	data, err := storage.GetByStatus(ctx, domain.OrderStratusProcessed)
	require.NoError(t, err)
	require.Empty(t, data)

	orderData.UserID = user2ID
	orderData.Number = "2345671"
	err = storage.Upload(ctx, orderData)
	require.NoError(t, err)

	data, err = storage.GetByStatus(ctx, domain.OrderStratusProcessing)
	require.NoError(t, err)
	require.Equal(t, 1, len(data)) // user1ID

	data, err = storage.GetByStatus(ctx, domain.OrderStratusNew)
	require.NoError(t, err)
	require.Equal(t, 1, len(data)) // user2ID

	ordData1 := domain.OrderData{
		Number:  "2345671",
		Status:  domain.OrderStratusProcessed,
		Accrual: domain.Float64Ptr(64.),
	}

	ordData2 := domain.OrderData{
		Number:  "1234561",
		Status:  domain.OrderStratusProcessed,
		Accrual: domain.Float64Ptr(65.),
	}

	err = storage.UpdateBatch(ctx, []domain.OrderData{ordData1, ordData2})
	require.NoError(t, err)
}

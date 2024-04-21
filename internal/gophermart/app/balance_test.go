package app_test

import (
	"context"
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/config"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/app"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/app/mocks"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestBalanceNoErr(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBalanceStorage(ctrl)

	mockStorage.EXPECT().Balance(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, uID int) (*domain.UserBalance, error) {
		require.Equal(t, userID, uID)
		return &domain.UserBalance{
			UserID:  userID,
			Release: 1,
			Balance: domain.Balance{
				Current:   100.,
				Withdrawn: 1000.4,
			},
		}, nil
	}).Times(1)

	conf := &config.GophermartConfig{}

	balance := app.NewBalance(conf, mockStorage)

	res, err := balance.Get(ctx)
	require.NotNil(t, res)

	require.NoError(t, err)

	require.Equal(t, 100., res.Current)
	require.Equal(t, 1000.4, res.Withdrawn)
}

func TestBalanceErrUserIsNotAuthorized(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBalanceStorage(ctrl)

	mockStorage.EXPECT().Balance(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, uID int) (*domain.UserBalance, error) {
		return nil, nil
	}).Times(0)

	conf := &config.GophermartConfig{}

	balance := app.NewBalance(conf, mockStorage)

	res, err := balance.Get(ctx)
	require.Nil(t, res)

	require.ErrorIs(t, err, domain.ErrUserIsNotAuthorized)
}

func TestBalanceErrServerInternal(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBalanceStorage(ctrl)

	mockStorage.EXPECT().Balance(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, uID int) (*domain.UserBalance, error) {
		require.Equal(t, userID, uID)
		return nil, fmt.Errorf("any error")
	}).Times(1)

	conf := &config.GophermartConfig{}

	balance := app.NewBalance(conf, mockStorage)

	res, err := balance.Get(ctx)
	require.Nil(t, res)

	require.ErrorIs(t, err, domain.ErrServerInternal)
}

func TestWithdrawNoErr(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBalanceStorage(ctrl)

	mockStorage.EXPECT().Balance(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, uID int) (*domain.UserBalance, error) {
		require.Equal(t, userID, uID)
		return &domain.UserBalance{
			UserID:  userID,
			Release: 1,
			Balance: domain.Balance{
				Current:   100.,
				Withdrawn: 1000.4,
			},
		}, nil
	}).Times(1)

	mockStorage.EXPECT().UpdateBalanceByWithdraw(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, newBalance *domain.UserBalance, withdraw *domain.WithdrawData) error {

			require.NotNil(t, newBalance)
			require.Equal(t, userID, newBalance.UserID)
			require.Equal(t, 1, newBalance.Release)
			require.Equal(t, 50., newBalance.Balance.Current)
			require.Equal(t, 1050.4, newBalance.Balance.Withdrawn)

			require.NotNil(t, withdraw)
			require.Equal(t, domain.OrderNumber("5062821234567892"), withdraw.Order)
			require.Equal(t, 50., withdraw.Sum)
			return nil
		}).Times(1)

	conf := &config.GophermartConfig{}

	balance := app.NewBalance(conf, mockStorage)

	err = balance.Withdraw(ctx, &domain.WithdrawData{
		Order: "5062821234567892",
		Sum:   50.,
	})

	require.NoError(t, err)
}

func TestWithdrawErrServerInternal(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBalanceStorage(ctrl)

	mockStorage.EXPECT().Balance(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, uID int) (*domain.UserBalance, error) {
		require.Equal(t, userID, uID)
		return nil, fmt.Errorf("any")
	}).Times(1)

	conf := &config.GophermartConfig{}

	balance := app.NewBalance(conf, mockStorage)

	err = balance.Withdraw(ctx, &domain.WithdrawData{
		Order: "5062821234567892",
		Sum:   50.,
	})

	require.ErrorIs(t, err, domain.ErrServerInternal)
}

func TestWithdrawErrUserIsNotAuthorized(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBalanceStorage(ctrl)

	conf := &config.GophermartConfig{}

	balance := app.NewBalance(conf, mockStorage)

	err := balance.Withdraw(ctx, &domain.WithdrawData{
		Order: "5062821234567892",
		Sum:   50.,
	})

	require.ErrorIs(t, err, domain.ErrUserIsNotAuthorized)
}

func TestWithdrawErrNotEnoughPoints(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBalanceStorage(ctrl)

	mockStorage.EXPECT().Balance(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, uID int) (*domain.UserBalance, error) {
		require.Equal(t, userID, uID)
		return &domain.UserBalance{
			UserID:  userID,
			Release: 1,
			Balance: domain.Balance{
				Current:   100.,
				Withdrawn: 1000.4,
			},
		}, nil
	}).Times(1)

	conf := &config.GophermartConfig{}

	balance := app.NewBalance(conf, mockStorage)

	err = balance.Withdraw(ctx, &domain.WithdrawData{
		Order: "5062821234567892",
		Sum:   500.,
	})

	require.ErrorIs(t, err, domain.ErrNotEnoughPoints)
}

func TestWithdrawErrWrongOrderNumber(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBalanceStorage(ctrl)

	conf := &config.GophermartConfig{}

	balance := app.NewBalance(conf, mockStorage)

	err = balance.Withdraw(ctx, &domain.WithdrawData{
		Order: "506282134567892",
		Sum:   500.,
	})

	require.ErrorIs(t, err, domain.ErrWrongOrderNumber)
}

func TestWithdrawalsNoErr(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBalanceStorage(ctrl)

	rTime := domain.RFC3339Time(time.Now())

	resData := []domain.WithdrawalData{
		{
			Order:       "5062821234567892",
			Sum:         50.,
			ProcessedAt: rTime,
		},
		{
			Order:       "5062821234567892",
			Sum:         150.,
			ProcessedAt: rTime,
		},
	}

	mockStorage.EXPECT().Withdrawals(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, uID int) ([]domain.WithdrawalData, error) {
		require.Equal(t, userID, uID)
		return resData, nil
	}).Times(1)

	conf := &config.GophermartConfig{}

	balance := app.NewBalance(conf, mockStorage)

	res, err := balance.Withdrawals(ctx)

	require.NoError(t, err)

	require.True(t, reflect.DeepEqual(res, resData))
}

func TestWithdrawalsErrServerInternal(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBalanceStorage(ctrl)

	mockStorage.EXPECT().Withdrawals(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, uID int) ([]domain.WithdrawalData, error) {
		require.Equal(t, userID, uID)
		return nil, fmt.Errorf("Any")
	}).Times(1)

	conf := &config.GophermartConfig{}

	balance := app.NewBalance(conf, mockStorage)

	res, err := balance.Withdrawals(ctx)

	require.Nil(t, res)
	require.ErrorIs(t, err, domain.ErrServerInternal)
}

func TestWithdrawalsErrUserIsNotAuthorized(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBalanceStorage(ctrl)

	conf := &config.GophermartConfig{}

	balance := app.NewBalance(conf, mockStorage)

	res, err := balance.Withdrawals(ctx)

	require.Nil(t, res)
	require.ErrorIs(t, err, domain.ErrUserIsNotAuthorized)
}

func TestWithdrawalsErrNotFound(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBalanceStorage(ctrl)

	mockStorage.EXPECT().Withdrawals(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, uID int) ([]domain.WithdrawalData, error) {
		require.Equal(t, userID, uID)
		return nil, nil
	}).Times(1)

	conf := &config.GophermartConfig{}

	balance := app.NewBalance(conf, mockStorage)

	res, err := balance.Withdrawals(ctx)

	require.Nil(t, res)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestPoolOrders(t *testing.T) {

	// Тест на отстуствие данных для обновления баланса
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	ctx = EnrichTestContext(ctx)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBalanceStorage(ctrl)

	mockStorage.EXPECT().GetByStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, status domain.OrderStatus) ([]domain.OrderData, error) {
			require.Equal(t, domain.OrderStratusProcessing, status)
			return nil, nil
		}).MinTimes(5).MaxTimes(6) // Ожидаем 10 секунд; 2 секунды между вызовами

	conf := &config.GophermartConfig{}

	bl := app.NewBalance(conf, mockStorage)

	bl.PoolOrders(ctx)

	time.Sleep(10 * time.Second)

	cancelFn()
}

func TestPoolOrders2(t *testing.T) {

	// Тест на отстуствие данных для пула системы расчета начислений

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	ctx = EnrichTestContext(ctx)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBalanceStorage(ctrl)

	userID := 100
	accrual := domain.Float64Ptr(105.)
	orderData := domain.OrderData{
		UserID:  userID,
		Status:  domain.OrderStratusProcessing,
		Accrual: accrual,
		Number:  "1234",
	}

	var once atomic.Int32

	mockStorage.EXPECT().GetByStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, status domain.OrderStatus) ([]domain.OrderData, error) {
			require.Equal(t, domain.OrderStratusProcessing, status)
			if once.CompareAndSwap(0, 1) { // Возвращаем данные только 1 раз
				return []domain.OrderData{
					orderData,
				}, nil
			}
			return nil, nil
		}).
		// Первый вызов вернул данные => второе обращение будет сразу за первым
		// второй вызов не венул данные => перед третьим обращением будет пауза в 2 секунды; всего в тесте ждем 3 секунды
		MinTimes(2).
		MaxTimes(3)

	mockStorage.EXPECT().Balance(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, userID int) (*domain.UserBalance, error) {
			require.Equal(t, 100, userID)
			return &domain.UserBalance{
				BalanceId: 10,
				UserID:    100,
				Release:   0,
				Balance: domain.Balance{
					Current:   100,
					Withdrawn: 0,
				},
			}, nil
		}).Times(1)

	mockStorage.EXPECT().UpdateBalanceByOrder(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, balance *domain.UserBalance, order *domain.OrderData) error {
			require.NotNil(t, balance)
			require.NotNil(t, order)

			require.Equal(t, 10, balance.BalanceId)
			require.Equal(t, 100, balance.UserID)
			require.Equal(t, 0, balance.Release)

			require.Equal(t, 100.+*accrual, balance.Balance.Current)
			require.Equal(t, 0., balance.Balance.Withdrawn)

			require.Equal(t, domain.OrderNumber("1234"), order.Number)
			require.Equal(t, domain.OrderStratusProcessed, order.Status)
			return nil
		}).Times(1)

	conf := &config.GophermartConfig{}

	bl := app.NewBalance(conf, mockStorage)

	bl.PoolOrders(ctx)

	time.Sleep(3 * time.Second)

	cancelFn()
}

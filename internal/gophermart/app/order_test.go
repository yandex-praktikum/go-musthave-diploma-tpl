package app_test

import (
	"context"
	"errors"
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

func TestNewNoErr(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	number := domain.OrderNumber("5062821234567892")

	mockStorage.EXPECT().Upload(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, oData *domain.OrderData) error {
		require.NotNil(t, oData)
		require.Equal(t, userID, oData.UserID)
		require.Equal(t, number, oData.Number)
		require.Equal(t, domain.OrderStratusNew, oData.Status)
		require.Nil(t, oData.Accrual)
		return nil
	})

	conf := &config.GophermartConfig{
		AcrualSystemPoolCount: 5,
	}

	order := app.NewOrder(ctx, conf, mockStorage, nil)

	err = order.New(ctx, number)

	require.NoError(t, err)
}

func TestNewErrOrderNumberAlreadyProcessed(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	number := domain.OrderNumber("5062821234567892")

	mockStorage.EXPECT().Upload(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, oData *domain.OrderData) error {
		require.NotNil(t, oData)
		require.Equal(t, userID, oData.UserID)
		require.Equal(t, number, oData.Number)
		require.Equal(t, domain.OrderStratusNew, oData.Status)
		require.Nil(t, oData.Accrual)
		return domain.ErrOrderNumberAlreadyUploaded
	})

	conf := &config.GophermartConfig{
		AcrualSystemPoolCount: 5,
	}

	order := app.NewOrder(ctx, conf, mockStorage, nil)

	err = order.New(ctx, number)

	require.ErrorIs(t, err, domain.ErrOrderNumberAlreadyUploaded)
}

func TestNewErrDublicateOrderNumber(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	number := domain.OrderNumber("5062821234567892")

	conf := &config.GophermartConfig{
		AcrualSystemPoolCount: 5,
	}

	order := app.NewOrder(ctx, conf, mockStorage, nil)

	mockStorage.EXPECT().Upload(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, oData *domain.OrderData) error {
		require.NotNil(t, oData)
		require.Equal(t, userID, oData.UserID)
		require.Equal(t, number, oData.Number)
		require.Equal(t, domain.OrderStratusNew, oData.Status)
		require.Nil(t, oData.Accrual)
		return domain.ErrDublicateOrderNumber
	})

	err = order.New(ctx, number)

	require.ErrorIs(t, err, domain.ErrDublicateOrderNumber)
}

func TestNewErrUserIsNotAuthorized(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	number := domain.OrderNumber("5062821234567892")

	conf := &config.GophermartConfig{
		AcrualSystemPoolCount: 5,
	}

	order := app.NewOrder(ctx, conf, mockStorage, nil)

	err := order.New(ctx, number)

	require.ErrorIs(t, err, domain.ErrUserIsNotAuthorized)
}

func TestNewErrServerInternal(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	number := domain.OrderNumber("5062821234567892")

	mockStorage.EXPECT().Upload(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, oData *domain.OrderData) error {
		return errors.New("err")
	})

	conf := &config.GophermartConfig{
		AcrualSystemPoolCount: 5,
	}

	order := app.NewOrder(ctx, conf, mockStorage, nil)

	err = order.New(ctx, number)

	require.ErrorIs(t, err, domain.ErrServerInternal)
}

func TestNewErrWrongOrderNumber(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	number := domain.OrderNumber("50628212345678921")

	conf := &config.GophermartConfig{
		AcrualSystemPoolCount: 5,
	}

	order := app.NewOrder(ctx, conf, mockStorage, nil)

	err = order.New(ctx, number)

	require.ErrorIs(t, err, domain.ErrWrongOrderNumber)
}

func TestAllErrNoErr(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	ordrs := []domain.OrderData{
		{
			UserID:     userID,
			Number:     "5062821234567891",
			Status:     domain.OrderStratusNew,
			UploadedAt: domain.RFC3339Time(time.Now()),
		},
		{
			UserID:     userID,
			Number:     "5062821234567892",
			Status:     domain.OrderStratusProcessed,
			Accrual:    domain.Float64Ptr(10.),
			UploadedAt: domain.RFC3339Time(time.Now()),
		},
	}

	mockStorage.EXPECT().Orders(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, uID int) ([]domain.OrderData, error) {
		require.Equal(t, userID, uID)
		return ordrs, nil
	}).Times(1)

	conf := &config.GophermartConfig{
		AcrualSystemPoolCount: 5,
	}

	order := app.NewOrder(ctx, conf, mockStorage, nil)

	res, err := order.All(ctx)

	require.Nil(t, err)

	require.True(t, reflect.DeepEqual(ordrs, res))
}

func TestAllErrUserIsNotAuthorized(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	conf := &config.GophermartConfig{
		AcrualSystemPoolCount: 5,
	}

	order := app.NewOrder(ctx, conf, mockStorage, nil)

	res, err := order.All(ctx)

	require.Nil(t, res)
	require.ErrorIs(t, err, domain.ErrUserIsNotAuthorized)
}

func TestAllErrNotFound(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	mockStorage.EXPECT().Orders(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, uID int) ([]domain.OrderData, error) {
		require.Equal(t, userID, uID)
		return nil, nil
	})
	conf := &config.GophermartConfig{
		AcrualSystemPoolCount: 5,
	}

	order := app.NewOrder(ctx, conf, mockStorage, nil)

	res, err := order.All(ctx)

	require.Nil(t, res)

	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestAllErrServerInternal(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	mockStorage.EXPECT().Orders(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, uID int) ([]domain.OrderData, error) {
		require.Equal(t, userID, uID)
		return nil, errors.New("any err")
	})

	conf := &config.GophermartConfig{
		AcrualSystemPoolCount: 5,
	}

	order := app.NewOrder(ctx, conf, mockStorage, nil)

	res, err := order.All(ctx)

	require.Nil(t, res)

	require.ErrorIs(t, err, domain.ErrServerInternal)
}

func TestPoolAcrualSystem1(t *testing.T) {

	// Тест на отстуствие данных для пула системы расчета начислений

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	ctx = EnrichTestContext(ctx)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	mockStorage.EXPECT().GetByStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, statuses []domain.OrderStatus) ([]domain.OrderData, error) {
			return nil, nil
		}).MinTimes(5).MaxTimes(6) // Ожидаем 10 секунд; 2 секунды между вызовами

	conf := &config.GophermartConfig{
		AcrualSystemPoolCount: 5,
	}

	order := app.NewOrder(ctx, conf, mockStorage, nil)

	order.PoolAcrualSystem(ctx)

	time.Sleep(10 * time.Second)
}

func TestPoolAcrualSystem2(t *testing.T) {

	// Тест на отстуствие данных для пула системы расчета начислений
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	ctx = EnrichTestContext(ctx)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	var once atomic.Int32

	ordNumber := domain.OrderNumber("12345")

	mockStorage.EXPECT().GetByStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, statuses []domain.OrderStatus) ([]domain.OrderData, error) {
			if once.CompareAndSwap(0, 1) {
				return []domain.OrderData{
					{
						Number: ordNumber,
					},
				}, nil
			}
			return nil, nil
		}).AnyTimes()

	mockAccrualSystem := mocks.NewMockAcrualSystem(ctrl)

	acrualVal := 64.

	mockAccrualSystem.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, orderNum domain.OrderNumber) (*domain.AccrualData, error) {
			return &domain.AccrualData{
				Number:  orderNum,
				Accrual: domain.Float64Ptr(acrualVal),
				Status:  domain.AccrualStatusProcessed,
			}, nil
		}).Times(1)

	mockStorage.EXPECT().UpdateBatch(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, orders []domain.OrderData) error {
			require.Equal(t, 1, len(orders))
			order := orders[0]
			require.Equal(t, ordNumber, order.Number)
			require.NotNil(t, order.Accrual)
			require.Equal(t, acrualVal, *order.Accrual)
			require.Equal(t, domain.OrderStratusProcessing, order.Status)
			return nil
		}).Times(1)

	conf := &config.GophermartConfig{
		AcrualSystemPoolCount: 5,
	}

	order := app.NewOrder(ctx, conf, mockStorage, mockAccrualSystem)

	order.PoolAcrualSystem(ctx)

	time.Sleep(10 * time.Second)
	cancelFn()
}

package app_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

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

	mockStorage.EXPECT().Balance(gomock.Any()).DoAndReturn(func(uID int) (*domain.UserBalance, error) {
		require.Equal(t, userID, uID)
		return &domain.UserBalance{
			UserID: userID,
			Score:  1,
			Balance: domain.Balance{
				Current:   100.,
				Withdrawn: 1000.4,
			},
		}, nil
	}).Times(1)

	balance := app.NewBalance(mockStorage)

	res, err := balance.Balance(ctx)
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

	mockStorage.EXPECT().Balance(gomock.Any()).DoAndReturn(func(uID int) (*domain.UserBalance, error) {
		return nil, nil
	}).Times(0)

	balance := app.NewBalance(mockStorage)

	res, err := balance.Balance(ctx)
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

	mockStorage.EXPECT().Balance(gomock.Any()).DoAndReturn(func(uID int) (*domain.UserBalance, error) {
		require.Equal(t, userID, uID)
		return nil, fmt.Errorf("any error")
	}).Times(1)

	balance := app.NewBalance(mockStorage)

	res, err := balance.Balance(ctx)
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

	mockStorage.EXPECT().Balance(gomock.Any()).DoAndReturn(func(uID int) (*domain.UserBalance, error) {
		require.Equal(t, userID, uID)
		return &domain.UserBalance{
			UserID: userID,
			Score:  1,
			Balance: domain.Balance{
				Current:   100.,
				Withdrawn: 1000.4,
			},
		}, nil
	}).Times(1)

	mockStorage.EXPECT().Withdraw(gomock.Any(), gomock.Any()).DoAndReturn(func(newBalance *domain.UserBalance, withdraw *domain.WithdrawData) error {

		require.NotNil(t, newBalance)
		require.Equal(t, userID, newBalance.UserID)
		require.Equal(t, 2, newBalance.Score)
		require.Equal(t, 50., newBalance.Balance.Current)
		require.Equal(t, 1050.4, newBalance.Balance.Withdrawn)

		require.NotNil(t, withdraw)
		require.Equal(t, domain.OrderNumber("5062821234567892"), withdraw.Order)
		require.Equal(t, 50., withdraw.Sum)
		return nil
	}).Times(1)

	balance := app.NewBalance(mockStorage)

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

	mockStorage.EXPECT().Balance(gomock.Any()).DoAndReturn(func(uID int) (*domain.UserBalance, error) {
		require.Equal(t, userID, uID)
		return nil, fmt.Errorf("any")
	}).Times(1)

	balance := app.NewBalance(mockStorage)

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

	balance := app.NewBalance(mockStorage)

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

	mockStorage.EXPECT().Balance(gomock.Any()).DoAndReturn(func(uID int) (*domain.UserBalance, error) {
		require.Equal(t, userID, uID)
		return &domain.UserBalance{
			UserID: userID,
			Score:  1,
			Balance: domain.Balance{
				Current:   100.,
				Withdrawn: 1000.4,
			},
		}, nil
	}).Times(1)

	balance := app.NewBalance(mockStorage)

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

	balance := app.NewBalance(mockStorage)

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

	resData := []domain.WithdrawalsData{
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

	mockStorage.EXPECT().Withdrawals(gomock.Any()).DoAndReturn(func(uID int) ([]domain.WithdrawalsData, error) {
		require.Equal(t, userID, uID)
		return resData, nil
	}).Times(1)

	balance := app.NewBalance(mockStorage)

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

	mockStorage.EXPECT().Withdrawals(gomock.Any()).DoAndReturn(func(uID int) ([]domain.WithdrawalsData, error) {
		require.Equal(t, userID, uID)
		return nil, fmt.Errorf("Any")
	}).Times(1)

	balance := app.NewBalance(mockStorage)

	res, err := balance.Withdrawals(ctx)

	require.Nil(t, res)
	require.ErrorIs(t, err, domain.ErrServerInternal)
}

func TestWithdrawalsErrUserIsNotAuthorized(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBalanceStorage(ctrl)
	balance := app.NewBalance(mockStorage)

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

	mockStorage.EXPECT().Withdrawals(gomock.Any()).DoAndReturn(func(uID int) ([]domain.WithdrawalsData, error) {
		require.Equal(t, userID, uID)
		return nil, nil
	}).Times(1)

	balance := app.NewBalance(mockStorage)

	res, err := balance.Withdrawals(ctx)

	require.Nil(t, res)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

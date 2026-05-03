package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/KalessinD/gophermart/internal/common"
	"github.com/KalessinD/gophermart/internal/middleware"
	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/repositories/postgresql/mocks"
	"github.com/KalessinD/gophermart/internal/services"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var UserID = "some-user-id"

// Вспомогательная функция для добавления Claims в контекст.
func setClaimsToCtx(ctx context.Context, t *testing.T, userID string) context.Context {
	t.Helper()
	return context.WithValue(ctx, middleware.ClaimsKey, &common.Claims{UserID: userID})
}

func TestBalanceService_GetBalanceInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockSQLStorageInterface(ctrl)
	service := services.NewBalanceService(mockDB)

	t.Run("success with withdrawals", func(t *testing.T) {
		ctx := context.Background()
		userID := UserID
		ctx = setClaimsToCtx(ctx, t, userID)

		user := &models.User{
			ID:      userID,
			Balance: 1000, // 10.00
		}
		withdrawn := &models.Withdrawn{
			Sum: 500, // 5.00
		}

		mockDB.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil)
		mockDB.EXPECT().GetWithdrawn(gomock.Any(), userID).Return(withdrawn, nil)

		balance, err := service.GetBalanceInfo(ctx)
		require.NoError(t, err)
		require.Equal(t, models.Accrual(1000), balance.Balance)
		require.Equal(t, models.Accrual(500), balance.Withdrawn)
	})

	t.Run("success without withdrawals", func(t *testing.T) {
		ctx := context.Background()
		userID := "user-2"
		ctx = setClaimsToCtx(ctx, t, userID)

		user := &models.User{
			ID:      userID,
			Balance: 200,
		}

		mockDB.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil)
		mockDB.EXPECT().GetWithdrawn(gomock.Any(), userID).Return(nil, models.ErrWithdrawnNotFound)

		balance, err := service.GetBalanceInfo(ctx)
		require.NoError(t, err)
		require.Equal(t, models.Accrual(200), balance.Balance)
		require.Equal(t, models.Accrual(0), balance.Withdrawn)
	})

	t.Run("user not found error", func(t *testing.T) {
		ctx := context.Background()
		userID := "user-404"
		ctx = setClaimsToCtx(ctx, t, userID)

		mockDB.EXPECT().GetUserByID(gomock.Any(), userID).Return(nil, errors.New("db error"))

		_, err := service.GetBalanceInfo(ctx)
		require.Error(t, err)
	})
}

func TestBalanceService_Withdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockSQLStorageInterface(ctrl)
	service := services.NewBalanceService(mockDB)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		userID := UserID
		ctx = setClaimsToCtx(ctx, t, userID)

		// Валидный номер по Луну (пример: 49927398716)
		withdrawn := &models.Withdrawn{
			ID:  "49927398716",
			Sum: 100,
		}

		// Ожидаем, что UserID будет проставлен из контекста
		mockDB.EXPECT().AddWithdrawn(gomock.Any(), gomock.Cond(func(x any) bool {
			w, ok := x.(*models.Withdrawn)
			if !ok {
				return false
			}
			return w.UserID == userID && w.ID == "49927398716"
		})).Return(nil)

		err := service.Withdraw(ctx, withdrawn)
		require.NoError(t, err)
	})

	t.Run("invalid order format (Luhn check)", func(t *testing.T) {
		ctx := context.Background()
		userID := UserID
		ctx = setClaimsToCtx(ctx, t, userID)

		withdrawn := &models.Withdrawn{
			ID:  "12345", // Невалидный номер
			Sum: 100,
		}

		// БД не должна вызываться
		err := service.Withdraw(ctx, withdrawn)
		require.ErrorIs(t, err, models.ErrOrderWrongFormat)
	})

	t.Run("db error", func(t *testing.T) {
		ctx := context.Background()
		userID := UserID
		ctx = setClaimsToCtx(ctx, t, userID)

		withdrawn := &models.Withdrawn{
			ID:  "49927398716", // Валидный
			Sum: 100,
		}

		mockDB.EXPECT().AddWithdrawn(gomock.Any(), gomock.Any()).Return(errors.New("insufficient funds"))

		err := service.Withdraw(ctx, withdrawn)
		require.Error(t, err)
		require.Contains(t, err.Error(), "insufficient funds")
	})
}

func TestBalanceService_ListWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockSQLStorageInterface(ctrl)
	service := services.NewBalanceService(mockDB)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		userID := UserID
		ctx = setClaimsToCtx(ctx, t, userID)

		expectedList := models.WithdrawnList{
			{ID: "order1", Sum: 100},
			{ID: "order2", Sum: 200},
		}

		mockDB.EXPECT().ListWithdrawals(gomock.Any(), userID).Return(expectedList, nil)

		list, err := service.ListWithdrawals(ctx)
		require.NoError(t, err)
		require.Len(t, list, 2)
	})

	t.Run("empty list", func(t *testing.T) {
		ctx := context.Background()
		userID := "user-2"
		ctx = setClaimsToCtx(ctx, t, userID)

		mockDB.EXPECT().ListWithdrawals(gomock.Any(), userID).Return(models.WithdrawnList{}, nil)

		list, err := service.ListWithdrawals(ctx)
		require.NoError(t, err)
		require.Empty(t, list)
	})
}

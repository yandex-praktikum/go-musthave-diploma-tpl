package loyalty

import (
	"context"
	"errors"
	"testing"
	"time"

	"gmart/internal/domain"
	"gmart/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLoyalty_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockLoyaltyRepoIface(ctrl)
	// Создаем объект напрямую, так как NewLoyalty создает реальный репозиторий
	svc := &Loyalty{repo: mockRepo}

	userID := domain.UserID(1)

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			GetBalance(gomock.Any(), userID).
			Return(domain.Amount(10050), domain.Amount(5000), nil).
			Times(1)

		curr, with, err := svc.GetBalance(context.Background(), userID)

		require.NoError(t, err)
		assert.Equal(t, domain.Amount(10050), curr)
		assert.Equal(t, domain.Amount(5000), with)
	})
}

func TestLoyalty_Withdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockLoyaltyRepoIface(ctrl)
	svc := &Loyalty{repo: mockRepo}

	userID := domain.UserID(1)
	validOrder := domain.OrderNumber("2377225624") // Валидный по Луну
	invalidOrder := domain.OrderNumber("12345")    // Невалидный
	amount := domain.Amount(50000)

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			Withdraw(gomock.Any(), userID, validOrder, amount).
			Return(nil).
			Times(1)

		err := svc.Withdraw(context.Background(), userID, validOrder, amount)
		assert.NoError(t, err)
	})

	t.Run("Invalid Order Number (Luhn)", func(t *testing.T) {
		// Репозиторий НЕ должен вызываться
		err := svc.Withdraw(context.Background(), userID, invalidOrder, amount)
		assert.ErrorIs(t, err, ErrInvalidOrderNumber)
	})

	t.Run("Insufficient Funds from Repo", func(t *testing.T) {
		// Имитируем ошибку баланса из БД
		errFunds := errors.New("insufficient balance")
		mockRepo.EXPECT().
			Withdraw(gomock.Any(), userID, validOrder, amount).
			Return(errFunds).
			Times(1)

		err := svc.Withdraw(context.Background(), userID, validOrder, amount)
		assert.Error(t, err)
		assert.Equal(t, errFunds, err)
	})
}

func TestLoyalty_GetWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockLoyaltyRepoIface(ctrl)
	svc := &Loyalty{repo: mockRepo}

	userID := domain.UserID(1)

	t.Run("Success", func(t *testing.T) {
		expected := []dto.WithdrawalItem{
			{
				Order:       "2377225624",
				Sum:         50000,
				ProcessedAt: time.Now(),
			},
		}

		mockRepo.EXPECT().
			GetWithdrawals(gomock.Any(), userID).
			Return(expected, nil).
			Times(1)

		res, err := svc.GetWithdrawals(context.Background(), userID)

		require.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, domain.OrderNumber("2377225624"), res[0].Order)
	})
}

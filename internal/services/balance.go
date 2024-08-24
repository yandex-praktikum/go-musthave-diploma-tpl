package services

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/eac0de/gophermart/internal/custom_errors"
	"github.com/eac0de/gophermart/internal/models"
	"github.com/eac0de/gophermart/pkg/utils"
	"github.com/google/uuid"
)

type BalanceService struct {
	mu            sync.Mutex
	withdrawStore WithdrawStore
}

func NewBalanceService(withdrawStore WithdrawStore) *BalanceService {
	return &BalanceService{
		withdrawStore: withdrawStore,
	}
}

func (bs *BalanceService) GetUserWithdrawals(ctx context.Context, userID uuid.UUID) ([]*models.Withdraw, error) {
	return bs.withdrawStore.SelectUserWithdrawals(ctx, userID)
}

func (bs *BalanceService) CreateWithdraw(ctx context.Context, orderNumber string, sum float32, userID uuid.UUID) (*models.Withdraw, error) {
	if sum == 0 {
		return nil, custom_errors.NewErrorWithHttpStatus("sum cannot be 0", http.StatusPaymentRequired)
	}
	if orderNumber == "" {
		return nil, custom_errors.NewErrorWithHttpStatus("order number cannot be empty", http.StatusUnprocessableEntity)
	}
	if !utils.CheckLuhnAlg(orderNumber) {
		return nil, custom_errors.NewErrorWithHttpStatus("order number did not pass the Luhn algorithm check", http.StatusUnprocessableEntity)
	}
	bs.mu.Lock()
	defer bs.mu.Unlock()
	user, err := bs.withdrawStore.SelectUserByID(ctx, userID)
	if err != nil {
		return nil, err

	}
	if user.Balance < sum {
		return nil, custom_errors.NewErrorWithHttpStatus("not enough points", http.StatusPaymentRequired)
	}
	user.Balance -= sum
	user.Withdrawn += sum
	err = bs.withdrawStore.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	withdraw := &models.Withdraw{
		ID:          uuid.New(),
		Order:       orderNumber,
		Sum:         sum,
		ProcessedAt: time.Now(),
		UserID:      user.ID,
	}
	err = bs.withdrawStore.InsertWithdraw(ctx, withdraw)
	if err != nil {
		return nil, err
	}
	return withdraw, nil
}

package service

import (
	"context"
	"errors"
	"strings"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/models"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"
)

// BalanceService — баланс и списания.
type BalanceService struct {
	orderRepo      repository.OrderRepository
	withdrawalRepo repository.WithdrawalRepository
}

// NewBalanceService создаёт сервис баланса.
func NewBalanceService(orderRepo repository.OrderRepository, withdrawalRepo repository.WithdrawalRepository) *BalanceService {
	return &BalanceService{orderRepo: orderRepo, withdrawalRepo: withdrawalRepo}
}

// Balance — текущий баланс и сумма списаний за всё время.
type Balance struct {
	Current   int64 // доступно (начисления − списания)
	Withdrawn int64 // всего списано
}

// GetBalance возвращает current (начисления − списания) и withdrawn для пользователя.
func (s *BalanceService) GetBalance(ctx context.Context, userID int64) (*Balance, error) {
	accruals, err := s.orderRepo.GetTotalAccrualsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	withdrawn, err := s.withdrawalRepo.GetTotalWithdrawnByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &Balance{
		Current:   accruals - withdrawn,
		Withdrawn: withdrawn,
	}, nil
}

// Withdraw списывает sum баллов в счёт заказа order.
func (s *BalanceService) Withdraw(ctx context.Context, userID int64, order string, sum int64) error {
	order = strings.TrimSpace(order)

	if !luhnValid(order) {
		return &ErrValidation{Msg: "invalid order number format (Luhn)"}
	}

	err := s.withdrawalRepo.Withdraw(ctx, userID, order, sum)
	if err != nil {
		var insufficient *repository.ErrInsufficientFunds
		if errors.As(err, &insufficient) {
			return err
		}
		var dup *repository.ErrDuplicateWithdrawalOrder
		if errors.As(err, &dup) {
			return err
		}
		return err
	}
	return nil
}

// ListWithdrawals возвращает списания пользователя от новых к старым.
func (s *BalanceService) ListWithdrawals(ctx context.Context, userID int64) ([]*models.Withdrawal, error) {
	return s.withdrawalRepo.ListByUserID(ctx, userID)
}

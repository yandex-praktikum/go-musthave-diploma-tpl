package repository

//go:generate mockgen -source=withdrawal.go -destination=mock/mock_withdrawal_repository.go -package=mock

import (
	"context"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/models"
)

// WithdrawalRepository — работа с таблицей списаний.
type WithdrawalRepository interface {
	// Create создаёт запись о списании. При дубликате (user_id, order) — *ErrDuplicateWithdrawalOrder.
	Create(ctx context.Context, userID int64, order string, sum int64) error
	// GetTotalWithdrawnByUserID возвращает сумму всех списаний пользователя (для баланса: withdrawn).
	GetTotalWithdrawnByUserID(ctx context.Context, userID int64) (int64, error)
	// ListByUserID возвращает все списания пользователя от новых к старым (ORDER BY processed_at DESC).
	ListByUserID(ctx context.Context, userID int64) ([]*models.Withdrawal, error)
	// Withdraw атомарно проверяет баланс (начисления − списания >= sum), вставляет списание. Блокировка по user_id (advisory lock). ErrInsufficientFunds или *ErrDuplicateWithdrawalOrder.
	Withdraw(ctx context.Context, userID int64, order string, sum int64) error
}

// Package repository предоставляет интерфейсы для доступа к данным.
// Интерфейсы определены рядом с потребителем (сервисами) для соблюдения DIP.
package repository

import (
	"context"

	"go-musthave-diploma-tpl/internal/storage"
)

// Ошибки репозиториев (алиасы из storage для удобства)
var (
	ErrUserNotFound       = storage.ErrUserNotFound
	ErrUserAlreadyExists  = storage.ErrUserAlreadyExists
	ErrOrderNotFound      = storage.ErrOrderNotFound
	ErrOrderAlreadyExists = storage.ErrOrderAlreadyExists
)

// UserRepository определяет контракт для работы с пользователями.
type UserRepository interface {
	// Create создаёт нового пользователя.
	Create(ctx context.Context, login string, passwordHash []byte) (*storage.User, error)
	// GetByLogin находит пользователя по логину.
	GetByLogin(ctx context.Context, login string) (*storage.User, error)
	// GetByID находит пользователя по ID.
	GetByID(ctx context.Context, id int64) (*storage.User, error)
}

// OrderRepository определяет контракт для работы с заказами.
type OrderRepository interface {
	// Create создаёт новый заказ.
	Create(ctx context.Context, userID int64, number string) (*storage.Order, error)
	// GetByNumber находит заказ по номеру.
	GetByNumber(ctx context.Context, number string) (*storage.Order, error)
	// GetByUser находит все заказы пользователя.
	GetByUser(ctx context.Context, userID int64) ([]storage.Order, error)
	// GetByUserAndNumber находит заказ пользователя по номеру.
	GetByUserAndNumber(ctx context.Context, userID int64, number string) (*storage.Order, error)
	// UpdateStatus обновляет статус заказа.
	UpdateStatus(ctx context.Context, id int64, status string, accrual *float64) error
	// GetOrdersByStatus получает заказы по статусу.
	GetOrdersByStatus(ctx context.Context, status string, limit int) ([]storage.Order, error)
}

// BalanceRepository определяет контракт для работы с балансом.
type BalanceRepository interface {
	// GetBalance получает текущий баланс пользователя.
	GetBalance(ctx context.Context, userID int64) (*storage.UserBalance, error)
	// ProcessWithdraw атомарно обрабатывает списание средств.
	ProcessWithdraw(ctx context.Context, userID int64, orderNumber string, amount float64) error
	// GetWithdrawalsByUser получает все списания пользователя.
	GetWithdrawalsByUser(ctx context.Context, userID int64) ([]storage.Withdrawal, error)
	// ProcessOrderAccrual обрабатывает начисление по заказу.
	ProcessOrderAccrual(ctx context.Context, orderID int64, userID int64, status string, accrual *float64, orderNumber string) error
}

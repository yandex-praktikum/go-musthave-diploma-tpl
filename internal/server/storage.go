package server

import (
	"context"

	"github.com/vglushak/go-musthave-diploma-tpl/internal/models"
)

// Storage интерфейс для хранения данных системы лояльности
type Storage interface {
	// User methods
	CreateUser(ctx context.Context, login, passwordHash string) (*models.User, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)

	// Order methods
	CreateOrder(ctx context.Context, userID int64, number string) (*models.Order, error)
	GetOrderByNumber(ctx context.Context, number string) (*models.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error)
	GetOrdersByStatus(ctx context.Context, statuses []string) ([]models.Order, error)
	GetOrdersByStatusPaginated(ctx context.Context, statuses []string, limit, offset int) ([]models.Order, error)
	UpdateOrderStatus(ctx context.Context, number string, status string, accrual *float64) error

	// Balance methods
	GetBalance(ctx context.Context, userID int64) (*models.Balance, error)
	UpdateBalance(ctx context.Context, userID int64, current, withdrawn float64) error

	// Withdrawal methods
	CreateWithdrawal(ctx context.Context, userID int64, order string, sum float64) (*models.Withdrawal, error)
	GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]models.Withdrawal, error)

	// Transactional withdrawal - проверяет баланс и создает списание в одной транзакции
	ProcessWithdrawal(ctx context.Context, userID int64, order string, sum float64) (*models.Withdrawal, error)

	// Атомарное обновление статуса заказа и баланса пользователя
	UpdateOrderStatusAndBalance(ctx context.Context, orderNumber string, status string, accrual *float64, userID int64, newCurrent, withdrawn float64) error

	// Database methods
	Ping(ctx context.Context) error
	Close() error
}

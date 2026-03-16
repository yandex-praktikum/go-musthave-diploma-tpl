package storage

import (
	"context"
	"errors"

	"github.com/NailUsmanov/internal/models"
	"go.uber.org/zap"
)

var ErrOrderAlreadyUsed = errors.New("order number already used")
var ErrOrderAlreadyUploaded = errors.New("order already uploaded by another person")
var ErrNotEnoughFunds = errors.New("insufficient funds")

type Auth interface {
	Registration(ctx context.Context, login string, password string) error
	GetUserByLogin(ctx context.Context, login string) (string, error)
	GetUserIDByLogin(ctx context.Context, login string) (int, error)
	CheckHashMatch(ctx context.Context, login, password string) error
}

type OrderOption interface {
	CreateNewOrder(ctx context.Context, userNumber int, numberOrder string, sugar *zap.SugaredLogger) error
	CheckExistOrder(ctx context.Context, numberOrder string) (bool, int, error)
	GetOrdersByUserID(ctx context.Context, userID int) ([]Order, error)
}

type WorkerAccrual interface {
	// GetOrdersForAccrualUpdate возвращает все заказы со статусами NEW, PROCESSING,
	// REGISTERED для обновления статуса и начислений
	GetOrdersForAccrualUpdate(ctx context.Context) ([]Order, error)

	// UpdateOrderStatus обновляет статус и сумму начислений по номеру заказа
	// (используется воркером после запроса к accrual-системе)
	UpdateOrderStatus(ctx context.Context, number string, status string, accrual *float64) error
}
type BalanceIndicator interface {
	// Для показаний текущего баланса и трат предыдущих
	GetUserBalance(ctx context.Context, userID int) (float64, float64, error)
	// Нахождение трат пользователя
	GetUserWithDrawns(ctx context.Context, userID int) (float64, error)
	// Добавление суммы списаний в таблицу с заказами
	AddWithdrawOrder(ctx context.Context, userID int, orderNumber string, sum float64) error
	// Вывод всех списаний конкретного пользователя
	GetAllUserWithdrawals(ctx context.Context, userID int) ([]models.UserWithDraw, error)
}

type Storage interface {
	Auth
	OrderOption
	WorkerAccrual
	BalanceIndicator
}

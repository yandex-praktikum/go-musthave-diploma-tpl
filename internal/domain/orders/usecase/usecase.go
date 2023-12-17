package usecase

import (
	"context"
	"database/sql"

	"github.com/benderr/gophermart/internal/domain/orders"
)

type OrderRepo interface {
	UpdateStatus(ctx context.Context, tx *sql.Tx, number string, status orders.Status) error
	UpdateAccrual(ctx context.Context, tx *sql.Tx, number string, accrual float64) error
	Create(ctx context.Context, userid string, number string, status orders.Status) (*orders.Order, error)
	GetByNumber(ctx context.Context, number string) (*orders.Order, error)
	GetOrdersByUser(ctx context.Context, userid string) ([]orders.Order, error)
}

type BalanceRepo interface {
	Add(ctx context.Context, tx *sql.Tx, userid string, balance float64) error
}

type Transactor interface {
	Within(ctx context.Context, tFunc func(ctx context.Context, tx *sql.Tx) error) error
}

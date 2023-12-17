package usecase

import (
	"context"
	"database/sql"

	"github.com/benderr/gophermart/internal/domain/balance"
)

type BalanceRepo interface {
	Add(ctx context.Context, tx *sql.Tx, userid string, balance float64) error
	GetBalanceByUser(ctx context.Context, tx *sql.Tx, userid string) (*balance.Balance, error)
	Withdraw(ctx context.Context, tx *sql.Tx, userid string, withdrawn float64) error
}

type WithdrawsRepo interface {
	Create(ctx context.Context, tx *sql.Tx, userid string, number string, sum float64) error
}

type Transactor interface {
	Within(ctx context.Context, tFunc func(ctx context.Context, tx *sql.Tx) error) error
}

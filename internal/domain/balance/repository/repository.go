package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/benderr/gophermart/internal/domain/balance"
	"github.com/benderr/gophermart/internal/logger"
)

type balanceRepository struct {
	db  *sql.DB
	log logger.Logger
}

func New(db *sql.DB, log logger.Logger) *balanceRepository {
	return &balanceRepository{db: db, log: log}
}

func (u *balanceRepository) GetBalanceByUser(ctx context.Context, tx *sql.Tx, userid string) (*balance.Balance, error) {

	row := tx.QueryRowContext(ctx, "SELECT user_id, current, withdrawn from balance WHERE user_id=$1", userid)
	var ord balance.Balance
	err := row.Scan(&ord.UserId, &ord.Current, &ord.Withdrawn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, balance.ErrNotFound
		}

		return nil, err
	}

	return &ord, nil
}

func (u *balanceRepository) Add(ctx context.Context, tx *sql.Tx, userid string, balance float64) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO balance (user_id, current)
	VALUES($1, $2) 
	ON CONFLICT (user_id) 
	DO UPDATE SET current=balance.current + $2`, userid, balance)

	return err
}

func (u *balanceRepository) Withdraw(ctx context.Context, tx *sql.Tx, userid string, withdrawn float64) error {
	_, err := tx.ExecContext(ctx, `UPDATE balance SET accrual=balance.withdrawn - $1 WHERE user_id=$2`, withdrawn, userid)
	return err
}

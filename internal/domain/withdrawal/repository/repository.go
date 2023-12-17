package repository

import (
	"context"
	"database/sql"

	"github.com/benderr/gophermart/internal/domain/withdrawal"
	"github.com/benderr/gophermart/internal/logger"
)

type withdrawalRepository struct {
	db  *sql.DB
	log logger.Logger
}

func New(db *sql.DB, log logger.Logger) *withdrawalRepository {
	return &withdrawalRepository{db: db, log: log}
}

func (w *withdrawalRepository) GetWithdrawsByUser(ctx context.Context, userid string) ([]withdrawal.Withdrawal, error) {
	list := make([]withdrawal.Withdrawal, 0)

	rows, err := w.db.QueryContext(ctx, "SELECT id, user_id, order_num, sum, processed_at from withdrawals WHERE user_id=$1 ORDER BY processed_at desc", userid)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var wdrl withdrawal.Withdrawal
		err = rows.Scan(&wdrl.Id, &wdrl.UserId, &wdrl.Order, &wdrl.Sum, &wdrl.PricessedAt)
		if err != nil {
			return nil, err
		}

		list = append(list, wdrl)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (u *withdrawalRepository) Create(ctx context.Context, tx *sql.Tx, userid string, order string, sum float64) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO withdrawals (user_id, order_num, sum) VALUES($1, $2, $3)`, userid, order, sum)
	return err
}

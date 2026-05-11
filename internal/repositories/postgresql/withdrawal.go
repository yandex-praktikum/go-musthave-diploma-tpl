package postgresql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/KalessinD/gophermart/internal/models"
)

const (
	PsqlWithdrawnTable = `"` + PsqlGophermartSchema + `"."withdrawals"`

	QueryInsertWithdrawn = `INSERT INTO ` + PsqlWithdrawnTable + ` AS t (user_id, order_id, withdrawn) VALUES($1, $2, $3) RETURNING id`

	QuerySelectWithdrawn = `SELECT user_id, SUM(withdrawn) FROM ` + PsqlWithdrawnTable + ` WHERE user_id = $1 GROUP BY user_id`

	QuerySelectWithdrawnList = `SELECT id, user_id, order_id, withdrawn, processed_at FROM ` + PsqlWithdrawnTable + ` WHERE user_id = $1`
)

func (r *SQLStorage) GetWithdrawn(ctx context.Context, userID string) (*models.Withdrawal, error) {
	wd := &models.Withdrawal{}

	_, err := r.withRetry(ctx, func(ctx context.Context) (*sql.Row, error) {
		row := r.db.QueryRowContext(ctx, QuerySelectWithdrawn, userID)
		if err := row.Scan(&wd.UserID, &wd.Sum); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, models.ErrWithdrawnNotFound
			}
			return nil, err
		}
		return row, nil
	})
	if err != nil {
		return nil, err
	}
	return wd, nil
}

func (r *SQLStorage) ListWithdrawals(ctx context.Context, userID string) (models.WithdrawalsList, error) {
	rows, err := r.db.QueryContext(ctx, QuerySelectWithdrawnList, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	wds := make(models.WithdrawalsList, 0, 100)

	for rows.Next() {
		wd := &models.Withdrawal{}
		err = rows.Scan(&wd.ID, &wd.UserID, &wd.OrderID, &wd.Sum, &wd.ProcessedAt)

		if err == nil {
			wds = append(wds, wd)
		} else {
			return nil, err
		}
	}

	return wds, nil
}

func (r *SQLStorage) AddWithdrawn(ctx context.Context, withdrawn *models.Withdrawal) error {
	return r.withTxRetry(ctx, func(tx *sql.Tx) error {
		if err := r.addWithdrawnTx(ctx, tx, withdrawn); err != nil {
			return err
		}
		if err := r.updateUserBalanceTx(ctx, tx, withdrawn.UserID, -withdrawn.Sum.Int()); err != nil {
			return err
		}
		return nil
	})
}

func (r *SQLStorage) addWithdrawnTx(ctx context.Context, tx *sql.Tx, wd *models.Withdrawal) error {
	err := tx.QueryRowContext(ctx, QueryInsertWithdrawn, wd.UserID, wd.OrderID, wd.Sum).Scan(&wd.ID)
	if err != nil {
		return err
	}
	return nil
}

package postgresql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/KalessinD/gophermart/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	PsqlOrderTable = `"` + PsqlGophermartSchema + `"."orders"`

	QueryInsertOrder = `INSERT INTO ` + PsqlOrderTable + ` AS t (id, user_id, status) VALUES($1, $2, $3) RETURNING id`

	QueryUpdateOrder = `
		INSERT INTO ` + PsqlOrderTable + ` AS t (id, user_id, status, accrual)
		VALUES($1, $2, $3, $4)
		ON CONFLICT (id)
		DO UPDATE SET
			accrual = EXCLUDED.accrual,
			status = EXCLUDED.status
		RETURNING id
	`

	QuerySelectOrder = `SELECT id, user_id, status, accrual, uploaded_at, updated_at FROM ` + PsqlOrderTable + ` WHERE id = $1 AND user_id = $2`

	QuerySelectOrders = `SELECT id, user_id, status, accrual, uploaded_at, updated_at FROM ` + PsqlOrderTable + ` WHERE user_id = $1`
)

func (r *SQLStorage) GetOrder(ctx context.Context, orderID, userID string) (*models.Order, error) {
	order := &models.Order{}

	_, err := r.withRetry(ctx, func(ctx context.Context) (*sql.Row, error) {
		row := r.db.QueryRowContext(ctx, QuerySelectOrder, orderID, userID)
		if err := row.Scan(&order.ID, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt, &order.UpdatedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, models.ErrOrderNotFound
			}
			return nil, err
		}
		return row, nil
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (r *SQLStorage) ListOrders(ctx context.Context, userID string) (models.OrdersList, error) {
	rows, err := r.db.QueryContext(ctx, QuerySelectOrders, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	orders := make(models.OrdersList, 0, 100)

	for rows.Next() {
		order := &models.Order{}
		err = rows.Scan(&order.ID, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt, &order.UpdatedAt)

		if err == nil {
			orders = append(orders, order)
		} else {
			return nil, err
		}
	}

	return orders, nil
}

func (r *SQLStorage) AddOrder(ctx context.Context, order *models.Order) error {
	return r.withTxRetry(ctx, func(tx *sql.Tx) error {
		if err := r.addOrderTx(ctx, tx, order); err != nil {
			return err
		}
		return nil
	})
}

func (r *SQLStorage) addOrderTx(ctx context.Context, tx *sql.Tx, order *models.Order) error {
	err := tx.QueryRowContext(ctx, QueryInsertOrder, order.ID, order.UserID, order.Status).Scan(&order.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return models.ErrOrderExists
			}
		}
		return err
	}
	return nil
}

func (r *SQLStorage) UpdateOrder(ctx context.Context, order *models.Order) error {
	return r.withTxRetry(ctx, func(tx *sql.Tx) error {
		if err := r.updateOrderTx(ctx, tx, order); err != nil {
			return err
		}
		if order.Status == models.OrderProcessedStatus && order.Accrual > 0 {
			if err := r.updateUserBalanceTx(ctx, tx, order.UserID, order.Accrual.Int()); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SQLStorage) updateOrderTx(ctx context.Context, tx *sql.Tx, order *models.Order) error {
	err := tx.QueryRowContext(ctx, QueryUpdateOrder, order.ID, order.UserID, order.Status, order.Accrual.Int()).Scan(&order.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return models.ErrOrderExists
			}
		}
		return err
	}
	return nil
}

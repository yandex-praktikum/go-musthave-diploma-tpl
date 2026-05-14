package repository

import (
	"context"
	"fmt"

	"github.com/AnxVit/go-musthave-diploma-tpl/internal/logger"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) GetNewOrders(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			order_number
		FROM orders
		WHERE status = ANY(ARRAY['NEW'::order_status, 'PROCESSING'::order_status]) AND withdrawn IS NULL AND is_registered = TRUE
	`)
	if err != nil {
		return nil, err
	}

	orders := make([]string, 0)
	for rows.Next() {
		var order string
		rows.Scan(&order)

		orders = append(orders, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, orders []model.AccuralResponse) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			logger.Log.Warn("Couldn't rollback", zap.Error(err))
		}
	}()

	for _, order := range orders {
		var (
			query string
			args  []interface{}
		)
		if order.Accrual == nil {
			query = `UPDATE orders 
					SET status = $2
                    WHERE order_number = $1`
			args = []interface{}{order.Order, order.Status}
		} else {
			query = `UPDATE orders
					SET 
						status = $2
						accrual = $3
                    WHERE order_number = $1`
			args = []interface{}{order.Order, order.Status, *order.Accrual}
		}

		_, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetNewUnregisteredOrders(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			order_number
		FROM orders
		WHERE status = 'NEW'::order_status AND withdrawn IS NULL AND is_registered = FALSE
	`)
	if err != nil {
		return nil, err
	}

	orders := make([]string, 0)
	for rows.Next() {
		var order string
		rows.Scan(&order)

		orders = append(orders, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *Repository) RegisterOrders(ctx context.Context, orders []string) error {
	if len(orders) == 0 {
		return fmt.Errorf("empty orders list")
	}
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE orders
		SET is_registered = TRUE
		WHERE order_number = ANY($1::text[])
	`, orders)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no orders update")
	}

	return nil
}

package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/A-Kuklin/gophermart/internal/accrual/types"
	"github.com/A-Kuklin/gophermart/internal/domain/entities"
)

const (
	getAccrualOrdersSQL = `
		SELECT id, status 
		FROM gophermart.orders
		WHERE status IN ($1, $2)
		ORDER BY updated_at DESC
		LIMIT $3
`
	updateAccrualOrdersSQL = `
		UPDATE gophermart.orders
		SET status = $1, accrual = $2
		WHERE id = $3
`
)

type AccrualPSQL struct {
	db *sql.DB
}

func NewAccrualPSQL(db *sql.DB) *AccrualPSQL {
	return &AccrualPSQL{
		db: db,
	}
}

func (s *AccrualPSQL) GetOrders(ctx context.Context, limit int) ([]types.Order, error) {
	stmt, err := s.db.PrepareContext(ctx, getAccrualOrdersSQL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var orders []types.Order
	rows, err := stmt.QueryContext(ctx, entities.StatusNew, entities.StatusProcessing, limit)
	for rows.Next() {
		var order types.Order
		err = rows.Scan(&order.OrderID, &order.Status)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *AccrualPSQL) UpdateOrderAccrual(ctx context.Context, orderAccrual types.AccrualResponse) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && err != nil {
			err = fmt.Errorf("UpdateOrderAccrual rollback error %s: %w", rollbackErr.Error(), err)
		}
	}()

	accrualCents := int64(orderAccrual.Accrual * 100)
	_, err = tx.ExecContext(ctx, updateAccrualOrdersSQL, orderAccrual.Status, accrualCents, orderAccrual.OrderID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

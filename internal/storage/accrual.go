package storage

// Accrual contract

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nikitaw13/go-musthave-diploma/internal/model"
)

// FetchNewOrders returns order numbers in NEW or PROCESSING status.
func (ps *PostgresStorage) FetchNewOrders(ctx context.Context) ([]string, error) {
	var result []string

	rows, err := ps.db.Query(ctx, fetchNewOrdersTemplate, model.StatusProcessing, model.StatusNew)

	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var orderNumber string
		err = rows.Scan(&orderNumber)
		if err != nil {
			return result, err
		}
		result = append(result, orderNumber)
	}

	err = rows.Err()
	if err != nil {
		return result, err
	}

	return result, nil
}

// UpdateOrder persists the accrual status and accrued amount of the order.
func (ps *PostgresStorage) UpdateOrder(ctx context.Context, orderNumber string, orderStatus string, accrual float64) error {
	operation := func() error {
		_, err := ps.db.Exec(ctx, updateOrderStatusTemplate, orderNumber, orderStatus,
			accrual, model.StatusProcessed, model.StatusInvalid)
		return err
	}

	err := ps.withRetries(ctx, operation)

	if err != nil {
		return err
	}
	return nil
}

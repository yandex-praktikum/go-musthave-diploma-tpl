package storage

// Service-order contract

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nikitaw13/go-musthave-diploma/internal/model"
)

// RegisterOrder inserts the order for the given user or, if it already exists,
// returns its current owner without modifying it.
func (ps *PostgresStorage) RegisterOrder(ctx context.Context, orderNumber, username string) (model.OrderRegistrationResult, error) {
	var result model.OrderRegistrationResult
	var insertedID int64

	operation := func() error {
		err := ps.db.QueryRow(ctx, insertOrderTemplate, orderNumber, username).Scan(&insertedID)
		return err
	}

	err := ps.withRetries(ctx, operation)
	if err == nil {
		result.Inserted, result.Owner = true, username
		return result, nil
	}

	// The insert was skipped because the order already exists; find its owner.
	if errors.Is(err, pgx.ErrNoRows) {
		operation := func() error {
			err := ps.db.QueryRow(ctx, getOrderOwnerTemplate, orderNumber).Scan(&result.Owner)
			return err
		}
		err := ps.withRetries(ctx, operation)

		if err != nil {
			result.Inserted, result.Owner = false, ""
			return result, err
		}
		result.Inserted = false
		return result, nil
	}
	return result, err
}

// ListOrders returns all orders of the given user.
func (ps *PostgresStorage) ListOrders(ctx context.Context, username string) ([]model.OrderResponse, error) {
	var result []model.OrderResponse

	operation := func() error {
		result = nil
		rows, err := ps.db.Query(ctx, listOrdersTemplate, username)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var row model.OrderResponse
			err = rows.Scan(&row.OrderNumber, &row.Status, &row.Amount, &row.UploadedTime)
			if err != nil {
				return err
			}
			result = append(result, row)
		}
		return rows.Err()
	}

	err := ps.withRetries(ctx, operation)

	if err != nil {
		return result, err
	}
	return result, nil
}

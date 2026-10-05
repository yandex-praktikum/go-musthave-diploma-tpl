package storage

// Service-withdrawal contract

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/nikitaw13/go-musthave-diploma/internal/model"
)

// GetBalance returns the current balance and total withdrawn of the given user.
func (ps *PostgresStorage) GetBalance(ctx context.Context, username string) (model.BalanceResponse, error) {
	var response model.BalanceResponse

	operation := func() error {
		return ps.db.QueryRow(ctx, getBalanceTemplate, username).Scan(&response.Current, &response.Withdrawn)
	}

	err := ps.withRetries(ctx, operation)
	if err != nil {
		return response, err
	}

	return response, nil
}

// Withdraw records a withdrawal of loyalty points for the given user or, if the order
// was already paid, returns its current owner without modifying it.
func (ps *PostgresStorage) Withdraw(ctx context.Context, username string, withdrawal model.WithdrawalRequest) (model.WithdrawalRegistrationResult, error) {
	var result model.WithdrawalRegistrationResult
	var response model.BalanceResponse
	var withdrawalID string

	operation := func() error {
		tx, err := ps.db.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		var userID int64
		err = tx.QueryRow(ctx, lockUserTemplate, username).Scan(&userID)
		if err != nil {
			return err
		}

		err = tx.QueryRow(ctx, getBalanceTemplate, username).Scan(&response.Current, &response.Withdrawn)
		if err != nil {
			return err
		}
		if response.Current < withdrawal.Amount {
			return model.ErrInsufficientFunds
		}
		err = tx.QueryRow(ctx, withdrawTemplate, username, withdrawal.OrderNumber, withdrawal.Amount).Scan(&withdrawalID)
		if err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	err := ps.withRetries(ctx, operation)
	if err == nil {
		result.Inserted, result.Owner = true, username
		return result, nil
	}

	// The insert was skipped because the withdrawal already exists; find its owner.
	if errors.Is(err, pgx.ErrNoRows) {
		operation := func() error {
			err := ps.db.QueryRow(ctx, getWithdrawalOwnerTemplate, withdrawal.OrderNumber).Scan(&result.Owner)
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

// ListWithdrawals returns the withdrawal history of the given user.
func (ps *PostgresStorage) ListWithdrawals(ctx context.Context, username string) ([]model.WithdrawalResponse, error) {
	var result []model.WithdrawalResponse

	operation := func() error {
		result = nil
		rows, err := ps.db.Query(ctx, listWithdrawalsTemplate, username)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var response model.WithdrawalResponse
			err = rows.Scan(&response.OrderNumber, &response.Amount, &response.ProcessedTime)
			if err != nil {
				return err
			}
			result = append(result, response)
		}
		return rows.Err()
	}

	err := ps.withRetries(ctx, operation)
	if err != nil {
		return result, err
	}

	return result, nil
}

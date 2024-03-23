package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/A-Kuklin/gophermart/internal/domain/entities"
)

const (
	createWithdrawSQL = `
		INSERT INTO gophermart.withdrawals (order_id, user_id, amount)
		VALUES ($1, $2, $3)
		RETURNING order_id, user_id, amount, created_at
`
	createWithdrawOrderSQL = `
		INSERT INTO gophermart.orders (id, user_id)
		VALUES ($1, $2)
`
	getSumWithdrawalsSQL = `
		SELECT SUM(amount)
		FROM gophermart.withdrawals
		WHERE user_id = $1
`
	getWithdrawalsSQL = `
		SELECT order_id, amount, created_at
		FROM gophermart.withdrawals
		WHERE user_id = $1
`
)

type WithdrawPSQL struct {
	db *sql.DB
}

func NewWithdrawPSQL(db *sql.DB) *WithdrawPSQL {
	return &WithdrawPSQL{
		db: db,
	}
}

func (s *WithdrawPSQL) CreateWithdraw(ctx context.Context, withdraw *entities.Withdraw) (*entities.Withdraw, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && err != nil {
			err = fmt.Errorf("CreateWithdraw rollback error %s: %w", rollbackErr.Error(), err)
		}
	}()

	_, err = tx.ExecContext(ctx, createWithdrawOrderSQL, withdraw.OrderID, withdraw.UserID)
	if err != nil {
		return nil, err
	}

	var withdrawDB entities.Withdraw
	result := tx.QueryRowContext(ctx, createWithdrawSQL, withdraw.OrderID, withdraw.UserID, withdraw.Amount)
	err = result.Scan(&withdrawDB.OrderID, &withdrawDB.UserID, &withdrawDB.Amount, &withdrawDB.CreatedAt)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &withdrawDB, nil
}

func (s *WithdrawPSQL) GetSumWithdrawals(ctx context.Context, userID uuid.UUID) (int64, error) {
	stmt, err := s.db.PrepareContext(ctx, getSumWithdrawalsSQL)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var withdrawalsCents sql.NullInt64
	result := stmt.QueryRowContext(ctx, userID)
	err = result.Scan(&withdrawalsCents)
	if err != nil {
		return 0, err
	}

	if !withdrawalsCents.Valid {
		withdrawalsCents.Int64 = 0
	}

	return withdrawalsCents.Int64, nil
}

func (s *WithdrawPSQL) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]entities.Withdraw, error) {
	stmt, err := s.db.PrepareContext(ctx, getWithdrawalsSQL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var withdrawals []entities.Withdraw
	rows, err := stmt.QueryContext(ctx, userID)
	for rows.Next() {
		var withdraw entities.Withdraw
		err = rows.Scan(&withdraw.OrderID, &withdraw.Amount, &withdraw.CreatedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdraw)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}

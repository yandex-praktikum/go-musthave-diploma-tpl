package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type BalanceRepository struct {
	db *sql.DB
}

func NewBalanceRepository(db *sql.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

func (r *BalanceRepository) GetAccrued(ctx context.Context, userID int64) (float64, error) {
	var accrued float64
	err := r.db.QueryRowContext(
		ctx,
		`SELECT COALESCE(SUM(accrual), 0)
		 FROM orders
		 WHERE user_id = $1 AND status = 'PROCESSED'`,
		userID,
	).Scan(&accrued)
	if err != nil {
		return 0, fmt.Errorf("get accrued: %w", err)
	}
	return accrued, nil
}

func (r *BalanceRepository) GetAccruedInTx(ctx context.Context, tx *sql.Tx, userID int64) (float64, error) {
	var accrued float64
	err := tx.QueryRowContext(
		ctx,
		`SELECT COALESCE(SUM(accrual), 0)
		 FROM orders
		 WHERE user_id = $1 AND status = 'PROCESSED'`,
		userID,
	).Scan(&accrued)
	if err != nil {
		return 0, fmt.Errorf("get accrued in tx: %w", err)
	}
	return accrued, nil
}

func (r *BalanceRepository) GetWithdrawn(ctx context.Context, userID int64) (float64, error) {
	var withdrawn float64
	err := r.db.QueryRowContext(
		ctx,
		`SELECT COALESCE(SUM(sum), 0)
		 FROM withdrawals
		 WHERE user_id = $1`,
		userID,
	).Scan(&withdrawn)
	if err != nil {
		return 0, fmt.Errorf("get withdrawn: %w", err)
	}
	return withdrawn, nil
}

func (r *BalanceRepository) GetWithdrawnInTx(ctx context.Context, tx *sql.Tx, userID int64) (float64, error) {
	var withdrawn float64
	err := tx.QueryRowContext(
		ctx,
		`SELECT COALESCE(SUM(sum), 0)
		 FROM withdrawals
		 WHERE user_id = $1`,
		userID,
	).Scan(&withdrawn)
	if err != nil {
		return 0, fmt.Errorf("get withdrawn in tx: %w", err)
	}
	return withdrawn, nil
}

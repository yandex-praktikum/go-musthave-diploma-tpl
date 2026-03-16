package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Withdrawal struct {
	Order       string
	Sum         float64
	ProcessedAt time.Time
}

type WithdrawalRepository struct {
	db *sql.DB
}

func NewWithdrawalRepository(db *sql.DB) *WithdrawalRepository {
	return &WithdrawalRepository{db: db}
}

func (r *WithdrawalRepository) Create(ctx context.Context, userID int64, order string, sum float64, processedAt time.Time) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO withdrawals (user_id, "order", sum, processed_at)
		 VALUES ($1, $2, $3, $4)`,
		userID, order, sum, processedAt,
	)
	if err != nil {
		return fmt.Errorf("create withdrawal: %w", err)
	}
	return nil
}

func (r *WithdrawalRepository) CreateInTx(ctx context.Context, tx *sql.Tx, userID int64, order string, sum float64, processedAt time.Time) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO withdrawals (user_id, "order", sum, processed_at)
		 VALUES ($1, $2, $3, $4)`,
		userID, order, sum, processedAt,
	)
	if err != nil {
		return fmt.Errorf("create withdrawal in tx: %w", err)
	}
	return nil
}

func (r *WithdrawalRepository) GetByUserID(ctx context.Context, userID int64) ([]Withdrawal, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT "order", sum, processed_at
		 FROM withdrawals
		 WHERE user_id = $1
		 ORDER BY processed_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get withdrawals by user id: %w", err)
	}
	defer rows.Close()

	var withdrawals []Withdrawal
	for rows.Next() {
		var w Withdrawal
		if err := rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, fmt.Errorf("scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return withdrawals, nil
}

func (r *WithdrawalRepository) GetTotalWithdrawn(ctx context.Context, userID int64) (float64, error) {
	var withdrawn float64
	err := r.db.QueryRowContext(
		ctx,
		`SELECT COALESCE(SUM(sum), 0)
		 FROM withdrawals
		 WHERE user_id = $1`,
		userID,
	).Scan(&withdrawn)
	if err != nil {
		return 0, fmt.Errorf("get total withdrawn: %w", err)
	}
	return withdrawn, nil
}

package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/models"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"
	"github.com/jackc/pgx/v5/pgconn"
)

// WithdrawalRepository — реализация WithdrawalRepository для PostgreSQL.
type WithdrawalRepository struct {
	db *sql.DB
}

// NewWithdrawalRepository создаёт репозиторий списаний.
func NewWithdrawalRepository(db *sql.DB) *WithdrawalRepository {
	return &WithdrawalRepository{db: db}
}

// Create вставляет запись о списании. При нарушении UNIQUE (user_id, order) — *repository.ErrDuplicateWithdrawalOrder.
func (r *WithdrawalRepository) Create(ctx context.Context, userID int64, order string, sum int64) error {
	q := `INSERT INTO withdrawals (user_id, "order", sum) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, q, userID, order, sum)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return &repository.ErrDuplicateWithdrawalOrder{Order: order}
		}
		return err
	}
	return nil
}

// GetTotalWithdrawnByUserID возвращает SUM(sum) по пользователю.
func (r *WithdrawalRepository) GetTotalWithdrawnByUserID(ctx context.Context, userID int64) (int64, error) {
	q := `SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1`
	var total int64
	err := r.db.QueryRowContext(ctx, q, userID).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// ListByUserID возвращает списания пользователя по processed_at DESC.
func (r *WithdrawalRepository) ListByUserID(ctx context.Context, userID int64) ([]*models.Withdrawal, error) {
	q := `SELECT id, user_id, "order", sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	return scanRows(rows, func(rows *sql.Rows) (*models.Withdrawal, error) {
		var w models.Withdrawal
		if err := rows.Scan(&w.ID, &w.UserID, &w.Order, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, err
		}
		return &w, nil
	})
}

// Withdraw атомарно: блокировка по user_id, проверка баланса (начисления − списания >= sum), вставка списания.
func (r *WithdrawalRepository) Withdraw(ctx context.Context, userID int64, order string, sum int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", userID)
	if err != nil {
		return err
	}

	var accruals int64
	err = tx.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(COALESCE(accrual, 0)), 0) FROM orders WHERE user_id = $1 AND status = 'PROCESSED'`,
		userID,
	).Scan(&accruals)
	if err != nil {
		return err
	}

	var withdrawn int64
	err = tx.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1`,
		userID,
	).Scan(&withdrawn)
	if err != nil {
		return err
	}

	if accruals-withdrawn < sum {
		return &repository.ErrInsufficientFunds{Order: order}
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO withdrawals (user_id, "order", sum) VALUES ($1, $2, $3)`, userID, order, sum)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return &repository.ErrDuplicateWithdrawalOrder{Order: order}
		}
		return err
	}
	return tx.Commit()
}

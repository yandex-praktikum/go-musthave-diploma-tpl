package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/models"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"
)

// OrderRepository — реализация OrderRepository для PostgreSQL.
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository создаёт репозиторий заказов.
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// nullInt64ToAccrual конвертирует sql.NullInt64 в *int для поля Accrual модели.
func nullInt64ToAccrual(n sql.NullInt64) *int {
	if !n.Valid {
		return nil
	}
	v := int(n.Int64)
	return &v
}

// GetByNumber возвращает заказ по номеру. Если не найден — *repository.ErrOrderNotFound.
func (r *OrderRepository) GetByNumber(ctx context.Context, number string) (*models.Order, error) {
	q := `SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE number = $1`
	var o models.Order
	var accrual sql.NullInt64
	err := r.db.QueryRowContext(ctx, q, number).Scan(
		&o.ID, &o.UserID, &o.Number, &o.Status, &accrual, &o.UploadedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &repository.ErrOrderNotFound{Number: number}
		}
		return nil, err
	}
	o.Accrual = nullInt64ToAccrual(accrual)
	return &o, nil
}

// Create вставляет заказ.
func (r *OrderRepository) Create(ctx context.Context, userID int64, number, status string) (*models.Order, error) {
	q := `INSERT INTO orders (user_id, number, status) VALUES ($1, $2, $3)
	      RETURNING id, user_id, number, status, accrual, uploaded_at`
	var o models.Order
	var accrual sql.NullInt64
	err := r.db.QueryRowContext(ctx, q, userID, number, status).Scan(
		&o.ID, &o.UserID, &o.Number, &o.Status, &accrual, &o.UploadedAt,
	)
	if err != nil {
		return nil, err
	}
	o.Accrual = nullInt64ToAccrual(accrual)
	return &o, nil
}

// ListByUserID возвращает заказы пользователя по uploaded_at DESC.
func (r *OrderRepository) ListByUserID(ctx context.Context, userID int64) ([]*models.Order, error) {
	q := `SELECT id, user_id, number, status, accrual, uploaded_at FROM orders
	      WHERE user_id = $1 ORDER BY uploaded_at DESC`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	return scanRows(rows, func(rows *sql.Rows) (*models.Order, error) {
		var o models.Order
		var accrual sql.NullInt64
		if err := rows.Scan(&o.ID, &o.UserID, &o.Number, &o.Status, &accrual, &o.UploadedAt); err != nil {
			return nil, err
		}
		o.Accrual = nullInt64ToAccrual(accrual)
		return &o, nil
	})
}

// UpdateAccrualAndStatus обновляет status и accrual заказа по номеру. Если заказ не найден — *repository.ErrOrderNotFound.
func (r *OrderRepository) UpdateAccrualAndStatus(ctx context.Context, number, status string, accrual *int) error {
	var accrualVal sql.NullInt64
	if accrual != nil {
		accrualVal = sql.NullInt64{Int64: int64(*accrual), Valid: true}
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`,
		status, accrualVal, number,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return &repository.ErrOrderNotFound{Number: number}
	}
	return nil
}

// ListNumbersPendingAccrual возвращает номера заказов в указанных статусах для опроса во внешней системе начислений.
func (r *OrderRepository) ListNumbersPendingAccrual(ctx context.Context, statuses []string) ([]string, error) {
	placeholders := make([]string, len(statuses))
	args := make([]interface{}, len(statuses))
	for i := range statuses {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = statuses[i]
	}
	q := `SELECT number FROM orders WHERE status IN (` + strings.Join(placeholders, ", ") + `) ORDER BY uploaded_at ASC`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	return scanRows(rows, func(rows *sql.Rows) (string, error) {
		var number string
		if err := rows.Scan(&number); err != nil {
			return "", err
		}
		return number, nil
	})
}

// GetTotalAccrualsByUserID возвращает SUM(COALESCE(accrual, 0)) по заказам пользователя со статусом PROCESSED.
func (r *OrderRepository) GetTotalAccrualsByUserID(ctx context.Context, userID int64) (int64, error) {
	q := `SELECT COALESCE(SUM(COALESCE(accrual, 0)), 0) FROM orders WHERE user_id = $1 AND status = 'PROCESSED'`
	var total int64
	err := r.db.QueryRowContext(ctx, q, userID).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

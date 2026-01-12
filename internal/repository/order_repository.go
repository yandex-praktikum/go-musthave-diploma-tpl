package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Order struct {
	ID         int64
	UserID     int64
	Number     string
	Status     string
	Accrual    *float64
	UploadedAt time.Time
}

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, userID int64, number string, status string, uploadedAt time.Time) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO orders (user_id, number, status, uploaded_at) VALUES ($1, $2, $3, $4)`,
		userID, number, status, uploadedAt,
	)
	if err != nil {
		return fmt.Errorf("create order: %w", err)
	}
	return nil
}

func (r *OrderRepository) GetByNumber(ctx context.Context, number string) (int64, error) {
	var userID int64
	err := r.db.QueryRowContext(
		ctx,
		`SELECT user_id FROM orders WHERE number = $1`,
		number,
	).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get order by number: %w", err)
	}
	return userID, nil
}

func (r *OrderRepository) GetByUserID(ctx context.Context, userID int64) ([]Order, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT number, status, accrual, uploaded_at
		 FROM orders
		 WHERE user_id = $1
		 ORDER BY uploaded_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get orders by user id: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var (
			number     string
			status     string
			accrual    sql.NullFloat64
			uploadedAt time.Time
		)
		if err := rows.Scan(&number, &status, &accrual, &uploadedAt); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		var accrualPtr *float64
		if accrual.Valid {
			v := accrual.Float64
			accrualPtr = &v
		}
		orders = append(orders, Order{
			UserID:     userID,
			Number:     number,
			Status:     status,
			Accrual:    accrualPtr,
			UploadedAt: uploadedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

type PendingOrder struct {
	ID     int64
	Number string
	UserID int64
}

func (r *OrderRepository) GetPendingOrders(ctx context.Context, limit int) ([]PendingOrder, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, number, user_id
		 FROM orders
		 WHERE status IN ('NEW', 'PROCESSING')
		 ORDER BY uploaded_at
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get pending orders: %w", err)
	}
	defer rows.Close()

	var orders []PendingOrder
	for rows.Next() {
		var o PendingOrder
		if err := rows.Scan(&o.ID, &o.Number, &o.UserID); err != nil {
			return nil, fmt.Errorf("scan pending order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE orders SET status = $1 WHERE id = $2`,
		status, orderID,
	)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return nil
}

func (r *OrderRepository) UpdateStatusWithAccrual(ctx context.Context, orderID int64, status string, accrual float64) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE orders
		 SET status = $1,
		     accrual = $2
		 WHERE id = $3`,
		status, accrual, orderID,
	)
	if err != nil {
		return fmt.Errorf("update order status with accrual: %w", err)
	}
	return nil
}

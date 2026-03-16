package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/models"
)

type OrderRepoPG struct {
	db *sql.DB
}

func NewOrderRepoPG(db *sql.DB) *OrderRepoPG {
	return &OrderRepoPG{db: db}
}

func (r *OrderRepoPG) CreateOrder(ctx context.Context, orderNumber string, userID int64) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO orders (order_number, user_id, status) VALUES ($1, $2, $3)`, orderNumber, userID, "NEW")
	return err
}

func (r *OrderRepoPG) GetOrderByNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	var o models.Order
	err := r.db.QueryRowContext(ctx, `SELECT id, order_number, user_id, status, created_at FROM orders WHERE order_number=$1`, orderNumber).Scan(&o.ID, &o.OrderNumber, &o.UserID, &o.Status, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepoPG) GetOrderByNumberAndUserID(ctx context.Context, orderNumber string, userID int64) (*models.Order, error) {
	var o models.Order
	err := r.db.QueryRowContext(ctx, `SELECT id, order_number, user_id, status, created_at FROM orders WHERE order_number=$1 AND user_id=$2`, orderNumber, userID).Scan(&o.ID, &o.OrderNumber, &o.UserID, &o.Status, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepoPG) GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, order_number, user_id, status, created_at FROM orders WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.OrderNumber, &o.UserID, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepoPG) GetOrdersForStatusUpdate(ctx context.Context) ([]models.Order, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, order_number, user_id, status, created_at FROM orders WHERE status IN ('NEW', 'PROCESSING')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.OrderNumber, &o.UserID, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepoPG) UpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE orders SET status=$1 WHERE id=$2`, status, orderID)
	return err
}

func (r *OrderRepoPG) AddBalanceTransaction(ctx context.Context, userID int64, orderID *int64, amount float64, txType string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO balance_transactions (user_id, order_id, amount, type) VALUES ($1, $2, $3, $4)`, userID, orderID, amount, txType)
	return err
}

func (r *OrderRepoPG) GetOrderAccrual(ctx context.Context, orderID int64) (*float64, error) {
	var accrual sql.NullFloat64
	err := r.db.QueryRowContext(ctx, `SELECT SUM(amount) FROM balance_transactions WHERE order_id=$1 AND type='ACCRUAL'`, orderID).Scan(&accrual)
	if err != nil {
		return nil, err
	}
	if !accrual.Valid {
		return nil, nil
	}
	return &accrual.Float64, nil
}

func (r *OrderRepoPG) GetUserBalance(ctx context.Context, userID int64) (current float64, withdrawn float64, err error) {
	err = r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(CASE WHEN type = 'ACCRUAL' THEN amount ELSE 0 END), 0) as accrual, COALESCE(SUM(CASE WHEN type = 'WITHDRAWAL' THEN amount ELSE 0 END), 0) as withdrawn FROM balance_transactions WHERE user_id = $1`, userID).Scan(&current, &withdrawn)
	if err != nil {
		return 0, 0, err
	}
	current = current - withdrawn
	return current, withdrawn, nil
}

func (r *OrderRepoPG) GetUserWithdrawals(ctx context.Context, userID int64) ([]models.WithdrawalResponse, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT order_id, amount, created_at FROM balance_transactions WHERE user_id=$1 AND type='WITHDRAWAL' ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var withdrawals []models.WithdrawalResponse
	for rows.Next() {
		var orderID sql.NullInt64
		var sum float64
		var processedAt sql.NullTime
		if err := rows.Scan(&orderID, &sum, &processedAt); err != nil {
			return nil, err
		}
		orderNumber := ""
		if orderID.Valid {
			var num string
			err = r.db.QueryRowContext(ctx, `SELECT order_number FROM orders WHERE id=$1`, orderID.Int64).Scan(&num)
			if err == nil {
				orderNumber = num
			}
		}
		withdrawals = append(withdrawals, models.WithdrawalResponse{
			Order:       orderNumber,
			Sum:         sum,
			ProcessedAt: processedAt.Time.Format(time.RFC3339),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return withdrawals, nil
}

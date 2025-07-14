package db

import (
	"database/sql"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/models"
)

type OrderRepoPG struct {
	db *sql.DB
}

func NewOrderRepoPG(db *sql.DB) *OrderRepoPG {
	return &OrderRepoPG{db: db}
}

func (r *OrderRepoPG) CreateOrder(orderNumber string, userID int64) error {
	_, err := r.db.Exec(`INSERT INTO orders (order_number, user_id, status) VALUES ($1, $2, $3)`, orderNumber, userID, "NEW")
	return err
}

func (r *OrderRepoPG) GetOrderByNumber(orderNumber string) (*models.Order, error) {
	var o models.Order
	err := r.db.QueryRow(`SELECT id, order_number, user_id, status, created_at FROM orders WHERE order_number=$1`, orderNumber).Scan(&o.ID, &o.OrderNumber, &o.UserID, &o.Status, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepoPG) GetOrderByNumberAndUserID(orderNumber string, userID int64) (*models.Order, error) {
	var o models.Order
	err := r.db.QueryRow(`SELECT id, order_number, user_id, status, created_at FROM orders WHERE order_number=$1 AND user_id=$2`, orderNumber, userID).Scan(&o.ID, &o.OrderNumber, &o.UserID, &o.Status, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepoPG) GetOrdersByUserID(userID int64) ([]models.Order, error) {
	rows, err := r.db.Query(`SELECT id, order_number, user_id, status, created_at FROM orders WHERE user_id=$1 ORDER BY created_at DESC`, userID)
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
	return orders, nil
}

func (r *OrderRepoPG) GetOrdersForStatusUpdate() ([]models.Order, error) {
	rows, err := r.db.Query(`SELECT id, order_number, user_id, status, created_at FROM orders WHERE status IN ('NEW', 'PROCESSING')`)
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
	return orders, nil
}

func (r *OrderRepoPG) UpdateOrderStatus(orderID int64, status string) error {
	_, err := r.db.Exec(`UPDATE orders SET status=$1 WHERE id=$2`, status, orderID)
	return err
}

func (r *OrderRepoPG) AddBalanceTransaction(userID int64, orderID *int64, amount float64, txType string) error {
	_, err := r.db.Exec(`INSERT INTO balance_transactions (user_id, order_id, amount, type) VALUES ($1, $2, $3, $4)`, userID, orderID, amount, txType)
	return err
}

func (r *OrderRepoPG) GetOrderAccrual(orderID int64) (*float64, error) {
	var accrual sql.NullFloat64
	err := r.db.QueryRow(`SELECT SUM(amount) FROM balance_transactions WHERE order_id=$1 AND type='ACCRUAL'`, orderID).Scan(&accrual)
	if err != nil {
		return nil, err
	}
	if !accrual.Valid {
		return nil, nil
	}
	return &accrual.Float64, nil
}

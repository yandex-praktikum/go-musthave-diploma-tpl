package repository

import (
	"gophermart/storage"
	"time"
)

const (
	NEW        = "NEW"
	PROCESSING = "PROCESSED"
	INVALID    = "INVALID"
	PROCESSED  = "PROCESSED"
)

type OrderRepository struct {
	DBStorage *storage.PgStorage
}

type OrderData struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float32   `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type UserBalance struct {
	Current   int `json:"current"`
	Withdrawn int `json:"withdrawn"`
}

func (or *OrderRepository) IsOrderExist(orderNumber string, userID int) (int, error) {
	var id, orderUserID int
	query := "SELECT id, user_id FROM orders WHERE number = $1"
	err := or.DBStorage.Conn.QueryRow(or.DBStorage.Ctx, query, orderNumber).Scan(&id, &orderUserID)

	if err != nil && err.Error() != "no rows in result set" {
		return 0, err
	}

	if err != nil && err.Error() == "no rows in result set" {
		return 0, nil
	}

	if orderUserID != userID {
		return 1, nil
	} else {
		return 2, nil
	}
}

func (or *OrderRepository) SaveOrder(orderNumber string, userID int) error {
	currentTime := time.Now()

	query := "INSERT INTO orders (number, user_id, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)"
	_, err := or.DBStorage.Conn.Exec(or.DBStorage.Ctx, query, orderNumber, userID, NEW, currentTime, currentTime)

	return err
}

func (or *OrderRepository) UpdateOrder(orderNumber string, accrual float32, status string) error {
	currentTime := time.Now()

	query := "UPDATE orders SET status = $1, accrual = $2, updated_at = $3 WHERE number = $4"
	_, err := or.DBStorage.Conn.Exec(or.DBStorage.Ctx, query, status, accrual, currentTime, orderNumber)

	return err
}

func (or *OrderRepository) GetUserOrders(userID int) ([]OrderData, error) {
	var orders []OrderData

	query := "SELECT number, status, accrual, created_at FROM orders WHERE user_id = $1 ORDER BY updated_at DESC"
	rows, err := or.DBStorage.Conn.Query(or.DBStorage.Ctx, query, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order OrderData
		if err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

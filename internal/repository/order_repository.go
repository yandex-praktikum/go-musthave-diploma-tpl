package repository

import (
	"github.com/sub3er0/gophermart/db"
	"time"
)

const (
	NEW        = "NEW"
	PROCESSING = "PROCESSED"
	INVALID    = "INVALID"
	PROCESSED  = "PROCESSED"
)

type OrderRepository struct {
	DBStorage *db.PgStorage
}

type OrderData struct {
	Number     int       `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type UserBalance struct {
	Balance     int `json:"balance"`
	UsedBalance int `json:"used_balance"`
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

func (or *OrderRepository) SaveOrder(orderNumber string, userID int, accrual int) error {
	query := "INSERT INTO orders (number, user_id, status, accrual) VALUES ($1, $2, $3, $4)"
	_, err := or.DBStorage.Conn.Exec(or.DBStorage.Ctx, query, orderNumber, userID, NEW, accrual)

	if accrual != 0 {
		query = "UPDATE user_balance SET balance = balance + $1 WHERE user_id = $2"
		_, err = or.DBStorage.Conn.Exec(or.DBStorage.Ctx, query, userID, accrual)
	}

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

func (or *OrderRepository) GetUserBalance(userID int) (UserBalance, error) {
	var userBalance UserBalance

	query := "SELECT balance, used_balance FROM user_balance WHERE user_id = $1"
	rows, err := or.DBStorage.Conn.Query(or.DBStorage.Ctx, query, userID)

	if err != nil {
		return userBalance, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&userBalance.Balance, &userBalance.UsedBalance); err != nil {
			return userBalance, err
		}
	}

	if err := rows.Err(); err != nil {
		return userBalance, err
	}

	return userBalance, nil
}

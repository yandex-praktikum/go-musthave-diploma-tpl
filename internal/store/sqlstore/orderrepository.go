package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/iRootPro/gophermart/internal/entity"
	"github.com/iRootPro/gophermart/internal/store"
	"time"
)

type OrderRepository struct {
	store *Store
}

func (o *OrderRepository) Create(order *entity.Order) error {
	foundOrder, err := o.FindByOrderNumber(order.Number)
	if err != nil {
		fmt.Println(err)
	}

	if foundOrder != nil && foundOrder.UserID == order.UserID {
		return store.ErrOrderNumberAlreadyExistInThisUser
	}

	if foundOrder != nil {
		return store.ErrOrderNumberAlreadyExistAnotherUser
	}

	var orderID int
	err = o.store.db.QueryRow("INSERT INTO orders (user_id, number, status, uploaded_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id", order.UserID, order.Number, "NEW", time.Now(), time.Now()).Scan(&orderID)
	if err != nil {
		return err
	}
	return nil
}

func (o *OrderRepository) FindByOrderNumber(id string) (*entity.Order, error) {
	order := &entity.Order{}
	err := o.store.db.QueryRow("SELECT id, user_id, number, status, uploaded_at, updated_at FROM orders WHERE number = $1", id).Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.UploadedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (o *OrderRepository) FindByUserID(userID int) (entity.Orders, error) {
	orders := entity.Orders{}
	rows, err := o.store.db.Query("SELECT id, user_id, number, status, uploaded_at, updated_at FROM orders WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		order := &entity.Order{}
		err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.UploadedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (o *OrderRepository) GetOrdersForUpgradeStatus() []string {
	rows, err := o.store.db.Query("SELECT number FROM orders WHERE status != $1 AND status != $2", "PROCESSED", "INVALID")
	if err != nil {
		fmt.Println(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(rows)

	var orders []string
	for rows.Next() {
		var number string
		err := rows.Scan(&number)
		if err != nil {
			fmt.Println(err)
		}
		orders = append(orders, number)
	}

	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}

	return orders
}

func (o *OrderRepository) UpdateStatus(order string, accrual float64, status string) (int, error) {
	var userID int
	err := o.store.db.QueryRow("UPDATE orders SET status = $1, accrual = accrual + $2, updated_at = $3 WHERE number = $3 RETURNING user_id", status, accrual, time.Now(), order).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

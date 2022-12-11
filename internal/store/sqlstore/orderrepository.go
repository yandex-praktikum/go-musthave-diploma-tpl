package sqlstore

import (
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

	fmt.Println("foundOrder", foundOrder)

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
	err := o.store.db.QueryRow("SELECT id, user_id, number, status, uploaded_at, updated_at FROM orders WHERE order_num = $1", id).Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.UploadedAt, &order.UpdatedAt)
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
	return orders, nil
}

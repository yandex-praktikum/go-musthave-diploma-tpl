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
	foundOrder, err := o.FindByOrderNumber(order.OrderNumber)
	if err != nil {
		fmt.Println(err)
	}

	if foundOrder.ID != 0 && foundOrder.UserID == order.UserID {
		return store.ErrOrderNumberAlreadyExistInThisUser
	}

	if foundOrder.ID != 0 {
		return store.ErrOrderNumberAlreadyExistAnotherUser
	}

	var orderID int
	err = o.store.db.QueryRow("INSERT INTO orders (user_id, order_num, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id", order.UserID, order.OrderNumber, "NEW", time.Now(), time.Now()).Scan(&orderID)
	if err != nil {
		return err
	}
	return nil
}

func (o *OrderRepository) FindByOrderNumber(id string) (*entity.Order, error) {
	order := &entity.Order{}
	err := o.store.db.QueryRow("SELECT id, user_id, order_num, status, created_at, updated_at FROM orders WHERE order_num = $1", id).Scan(&order.ID, &order.UserID, &order.OrderNumber, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return order, nil
}

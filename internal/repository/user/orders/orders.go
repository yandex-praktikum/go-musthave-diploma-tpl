package orders

import (
	"database/sql"
	"errors"

	"github.com/s-lyash/go-musthave-diploma-tpl/internal/models"
)

func (c client) GetUserOrder() {
	//TODO implement me
	panic("implement me")
}

func (c client) SaveUserOrder(order models.SaveOrder) int {
	var id string

	i := c.checkConflict(order)
	if i > 0 {
		return i
	}

	query := `INSERT INTO orders (user_id, order_id, status)
              VALUES ($1, $2, $3)
              ON CONFLICT (order_id) DO NOTHING
              RETURNING user_id`

	err := c.conn.QueryRow(query, order.UserID, order.OrderID, order.Status).Scan(&id)

	if err != nil && errors.Is(err, sql.ErrNoRows) {

		return models.InternalServerError
	}

	return models.Accepted
}

func (c client) checkConflict(order models.SaveOrder) int {
	var user, id string
	query := `SELECT user_id, order_id 
                  FROM orders 
                  WHERE order_id = $1`

	err := c.conn.QueryRow(query, order.OrderID).Scan(&user, &id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return 0
	}

	if user != order.UserID {
		return models.DownloadedByAnotherUser
	}

	return models.DownloadedByUser
}

func (c client) ChangeOrderStatus(order models.LoyaltySystem) {
	query := `UPDATE orders  
              SET status = $1 
              WHERE order_id = $2`

	err := c.conn.QueryRow(query, order.Status, order.Order)
	if err != nil {
		return
	}
	return
}

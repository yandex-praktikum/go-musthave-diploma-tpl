package orders

import (
	"database/sql"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/models"
)

type Repository interface {
	GetUserOrder()
	SaveUserOrder(models.SaveOrder) int
	ChangeOrderStatus(models.LoyaltySystem)
}

type client struct {
	conn *sql.DB
}

func NewRepository(conn *sql.DB) Repository {
	return client{
		conn: conn,
	}
}

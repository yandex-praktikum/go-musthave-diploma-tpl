package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
)

type Autorisation interface {
	CreateUser(user models.User) (int, error)
	GetUser(username string) (models.User, error)
}

type Orders interface {
	CreateOrder(num, user_id int, status string) (int, time.Time, error)
	GetOrders(user_id int) ([]models.Order, error)
}

type Repository struct {
	Autorisation
	Orders
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Autorisation: NewAuthPostgres(db),
		Orders:       NewOrdersPostgres(db),
	}
}

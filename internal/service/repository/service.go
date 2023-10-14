package repository

import (
	"database/sql"

	"github.com/s-lyash/go-musthave-diploma-tpl/internal/repository/user/auth"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/repository/user/balance"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/repository/user/orders"
)

type Repository struct {
	Auth    auth.Repository
	Orders  orders.Repository
	Balance balance.Repository
}

func NewRepository(conn *sql.DB) Repository {
	return Repository{
		Auth:    auth.NewRepository(conn),
		Orders:  orders.NewRepository(conn),
		Balance: balance.NewRepository(conn),
	}
}

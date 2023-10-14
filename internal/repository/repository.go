package repository

import (
	"database/sql"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/repository/user/auth"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/repository/user/balance"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/repository/user/orders"
)

type Conn struct {
	Conn     *sql.DB
	UserAuth auth.Repository
	Order    orders.Repository
	Balances balance.Repository
}

func NewDatabase(connectionData string) (*Conn, error) {
	db, err := sql.Open("pgx", connectionData)
	if err != nil {
		return nil, err
	}

	order := orders.NewRepository(db)
	user := auth.NewRepository(db)
	balances := balance.NewRepository(db)

	conn := &Conn{
		Conn:     db,
		UserAuth: user,
		Order:    order,
		Balances: balances,
	}

	return conn, err
}

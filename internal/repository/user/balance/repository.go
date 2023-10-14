package balance

import "database/sql"

type Repository interface {
	GetUserBalance()
	WithdrawUserBalance()
	GetUSerWithdrawals()
}

type client struct {
	conn *sql.DB
}

func NewRepository(conn *sql.DB) Repository {
	return client{
		conn: conn,
	}
}

package auth

import (
	"database/sql"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/models"
)

type Repository interface {
	GetUserID(models.User) (string, error)
	SaveUser(models.User) (string, error)
	FindUserByID(string) bool
}

type client struct {
	conn *sql.DB
}

func NewRepository(conn *sql.DB) Repository {
	return client{
		conn: conn,
	}
}

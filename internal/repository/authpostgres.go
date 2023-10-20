package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/models"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

func (r *AuthPostgres) CreateUser(user models.User) (int, error) {
	var id int

	query := fmt.Sprintf("INSERT INTO %s (login, password_hash, salt) values ($1, $2, $3) RETURNING id", usersTable)
	row := r.db.QueryRow(query, user.Login, user.Password, user.Salt)

	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *AuthPostgres) GetUser(username string) (models.User, error) {
	var user models.User
	query := fmt.Sprintf("SELECT id, login, password_hash, salt FROM %s WHERE login=$1", usersTable)
	err := r.db.Get(&user, query, username)

	return user, err
}

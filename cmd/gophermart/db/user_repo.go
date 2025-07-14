package db

import (
	"database/sql"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/models"
)

type UserRepoPG struct {
	db *sql.DB
}

func NewUserRepoPG(db *sql.DB) *UserRepoPG {
	return &UserRepoPG{db: db}
}

func (r *UserRepoPG) CreateUser(login, passwordHash string) error {
	_, err := r.db.Exec(`INSERT INTO users (login, password_hash) VALUES ($1, $2)`, login, passwordHash)
	return err
}

func (r *UserRepoPG) IsLoginExist(login string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE login=$1)`, login).Scan(&exists)
	return exists, err
}

func (r *UserRepoPG) GetUserByLogin(login string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(`SELECT id, login, password_hash, created_at FROM users WHERE login=$1`, login).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

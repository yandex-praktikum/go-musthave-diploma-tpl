package repository

import (
	"gophermart/db"
	"gophermart/internal/models"
)

type UserRepository struct {
	DBStorage *db.PgStorage
}

func (ur *UserRepository) CreateUser(user models.User) error {
	query := "INSERT INTO users (username, password) VALUES ($1, $2)"
	_, err := ur.DBStorage.Conn.Exec(ur.DBStorage.Ctx, query, user.Username, user.Password)

	return err
}

func (ur *UserRepository) CreateUserBalance(user models.User) error {
	var id int
	query := "SELECT id FROM users WHERE username = $1"
	err := ur.DBStorage.Conn.QueryRow(ur.DBStorage.Ctx, query, user.Username).Scan(&id)

	query = "INSERT INTO user_balance (user_id, balance) VALUES ($1, 0)"
	_, err = ur.DBStorage.Conn.Exec(ur.DBStorage.Ctx, query, id)

	return err
}

func (ur *UserRepository) IsUserExists(username string) int {
	var id int
	query := "SELECT id FROM users WHERE username = $1"
	err := ur.DBStorage.Conn.QueryRow(ur.DBStorage.Ctx, query, username).Scan(&id)

	if err != nil && err.Error() != "no rows in result set" {
		return -2
	}

	if err != nil && err.Error() == "no rows in result set" {
		return -1
	}

	return id
}

func (ur *UserRepository) GetUserByUsername(username string) (models.User, error) {
	var user models.User
	err := ur.DBStorage.Conn.QueryRow(
		ur.DBStorage.Ctx,
		"SELECT id, username, password FROM users WHERE username = $1", username).
		Scan(&user.ID, &user.Username, &user.Password)
	return user, err
}

package repository

import (
	"gophermart/internal/models"
	"gophermart/storage"
	"time"
)

type UserRepository struct {
	DBStorage *storage.PgStorage
}

func (ur *UserRepository) CreateUser(user models.User) (int, error) {
	currentTime := time.Now()
	query := "INSERT INTO users (username, password, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id"

	var userID int
	err := ur.DBStorage.Conn.QueryRow(ur.DBStorage.Ctx, query, user.Username, user.Password, currentTime, currentTime).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (ur *UserRepository) GetUserId(username string) int {
	var id int
	query := "SELECT id FROM users WHERE username = $1"
	err := ur.DBStorage.Conn.QueryRow(ur.DBStorage.Ctx, query, username).Scan(&id)

	if err != nil && err.Error() != "no rows in result set" {
		return DatabaseError
	}

	if err != nil && err.Error() == "no rows in result set" {
		return UserNotFound
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

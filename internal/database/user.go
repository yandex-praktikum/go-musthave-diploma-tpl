package database

import (
	"context"
	"fmt"

	"github.com/eac0de/gophermart/internal/models"
	"github.com/google/uuid"
)

func (db *Database) SelectUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := "SELECT id, name, age, username, balance, withdrawn, password FROM users WHERE username=$1"
	row := db.QueryRow(ctx, query, username)
	user := models.User{}
	err := row.Scan(&user.ID, &user.Name, &user.Age, &user.Username, &user.Balance, &user.Withdrawn, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *Database) SelectUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	query := "SELECT id, name, age, username, balance, withdrawn, password  FROM users WHERE id=$1"
	row := db.QueryRow(ctx, query, userID)
	user := models.User{}
	err := row.Scan(&user.ID, &user.Name, &user.Age, &user.Username, &user.Balance, &user.Withdrawn, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *Database) InsertUser(ctx context.Context, user *models.User) error {
	query := "INSERT INTO users (id, name, age, username, balance, withdrawn, password) VALUES($1,$2,$3,$4,$5,$6,$7)"
	_, err := db.Exec(ctx, query, user.ID, user.Name, user.Age, user.Username, user.Balance, user.Withdrawn, user.Password)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	query := "DELETE FROM users WHERE id=$1;"
	_, err := db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) UpdateUser(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		return fmt.Errorf("when updating user, his id can't be empty")
	}
	query := "UPDATE users SET name=$2, age=$3, username=$4, balance=$5, withdrawn=$6, password=$7, refresh_token=$8 WHERE id=$1"
	_, err := db.Exec(
		ctx,
		query,
		user.ID,
		user.Name,
		user.Age,
		user.Username,
		user.Balance,
		user.Withdrawn,
		user.Password,
		user.RefreshToken,
	)
	if err != nil {
		return err
	}
	return nil
}

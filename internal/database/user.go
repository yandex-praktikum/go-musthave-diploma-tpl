package database

import (
	"context"
	"fmt"

	"github.com/eac0de/gophermart/internal/models"
	"github.com/google/uuid"
)

func (db *Database) SelectUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := "SELECT id, name, age, email, email_confirmed, balance, password FROM users WHERE email=$1"
	row := db.QueryRowContext(ctx, query, email)
	user := models.User{}
	err := row.Scan(&user.ID, &user.Name, &user.Age, &user.Email, &user.EmailConfirmed, &user.Balance, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *Database) SelectUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	query := "SELECT id, name, age, email, email_confirmed, balance, password  FROM users WHERE id=$1"
	row := db.QueryRowContext(ctx, query, userID)
	user := models.User{}
	err := row.Scan(&user.ID, &user.Name, &user.Age, &user.Email, &user.EmailConfirmed, &user.Balance, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *Database) InsertUser(ctx context.Context, user *models.User) error {
	query := "INSERT INTO users (id, name, age, email, email_confirmed, balance, password) VALUES($1,$2,$3,$4,$5,$6,$7)"
	_, err := db.ExecContext(ctx, query, user.ID, user.Name, user.Age, user.Email, user.EmailConfirmed, user.Balance, user.Password)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	query := "DELETE FROM users WHERE id=$1;"
	_, err := db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) UpdateUser(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		return fmt.Errorf("when updating user, his id can't be empty")
	}
	query := "UPDATE users SET name=$2, age=$3, email=$4, email_confirmed=$5, balance=$6, password=$7, refresh_token=$8 WHERE id=$1"
	_, err := db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Name,
		user.Age,
		user.Email,
		user.EmailConfirmed,
		user.Balance,
		user.Password,
		user.RefreshToken,
	)
	if err != nil {
		return err
	}
	return nil
}

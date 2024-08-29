package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (db *Database) SaveRefreshToken(ctx context.Context, RefreshToken string, userID *uuid.UUID) error {
	query := "SELECT id FROM users WHERE id=$1"
	row := db.QueryRow(ctx, query, *userID)
	var existingID uuid.UUID
	err := row.Scan(&existingID)
	if err != nil {
		return fmt.Errorf("user with user id %v not found", *userID)
	}
	query = "UPDATE users SET refresh_token=$1 WHERE id=$2"
	_, err = db.Exec(ctx, query, RefreshToken, *userID)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) GetRefreshToken(ctx context.Context, userID *uuid.UUID) string {
	query := "SELECT refresh_token FROM users WHERE id=$1"
	row := db.QueryRow(ctx, query, *userID)
	var refreshToken string
	err := row.Scan(&refreshToken)
	if err != nil {
		return ""
	}
	return refreshToken
}

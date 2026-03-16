package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, login, passwordHash string) (int64, error) {
	var userID int64
	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`,
		login, passwordHash,
	).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}
	return userID, nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (int64, string, error) {
	var (
		userID       int64
		passwordHash string
	)
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, password_hash FROM users WHERE login = $1`,
		login,
	).Scan(&userID, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", ErrNotFound
	}
	if err != nil {
		return 0, "", fmt.Errorf("get user by login: %w", err)
	}
	return userID, passwordHash, nil
}

var ErrNotFound = errors.New("not found")

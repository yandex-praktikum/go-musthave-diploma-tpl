package storage

// Service-user contract

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nikitaw13/go-musthave-diploma/internal/model"
)

// UserExists reports whether a user with the given username exists.
func (ps *PostgresStorage) UserExists(ctx context.Context, username string) (bool, error) {
	var exists bool

	operation := func() error {
		err := ps.db.QueryRow(ctx, userExistsTemplate, username).Scan(&exists)
		return err
	}

	err := ps.withRetries(ctx, operation)

	if err != nil {
		return false, err
	}

	return exists, nil
}

// CreateUser inserts a new user with the given username and password hash.
func (ps *PostgresStorage) CreateUser(ctx context.Context, username string, passwordHash string) error {
	var pgErr *pgconn.PgError
	operation := func() error {
		_, err := ps.db.Exec(ctx, insertUserTemplate, username, passwordHash)
		return err
	}

	err := ps.withRetries(ctx, operation)

	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return model.ErrUserAlreadyExists
	}

	if err != nil {
		return err
	}

	return nil
}

// GetPasswordHash returns the stored password hash for the given username, or model.ErrUserNotFound.
func (ps *PostgresStorage) GetPasswordHash(ctx context.Context, username string) (string, error) {
	var passwordHash string

	operation := func() error {
		err := ps.db.QueryRow(ctx, getPasswordHashTemplate, username).Scan(&passwordHash)
		return err
	}

	err := ps.withRetries(ctx, operation)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", model.ErrUserNotFound
	}
	if err != nil {
		return "", err
	}
	return passwordHash, nil
}

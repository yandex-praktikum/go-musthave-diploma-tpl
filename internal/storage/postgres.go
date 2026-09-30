package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nikitaw13/go-musthave-diploma/internal/model"
	"go.uber.org/zap"
)

// PostgresStorage implements the service.Repository interface on top of PostgreSQL.
type PostgresStorage struct {
	db         *pgxpool.Pool
	timeouts   []time.Duration
	classifier *PostgresErrorClassifier
	logger     *zap.Logger
}

// NewPostgresStorage creates a PostgresStorage backed by the given pgxpool.
func NewPostgresStorage(db *pgxpool.Pool, timeouts []time.Duration, classifier *PostgresErrorClassifier, logger *zap.Logger) *PostgresStorage {
	return &PostgresStorage{
		db:         db,
		timeouts:   timeouts,
		classifier: classifier,
		logger:     logger,
	}
}

// UserExists reports whether a user with the given username exists.
func (ps *PostgresStorage) UserExists(ctx context.Context, username string) (bool, error) {
	var exists bool

	err := ps.db.QueryRow(ctx, userExistsTemplate, username).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetPasswordHash returns the stored password hash for the given username, or model.ErrUserNotFound.
func (ps *PostgresStorage) GetPasswordHash(ctx context.Context, username string) (string, error) {
	var passwordHash string

	row := ps.db.QueryRow(ctx, getPasswordHashTemplate, username)
	err := row.Scan(&passwordHash)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", model.ErrUserNotFound
	}

	if err != nil {
		return "", err
	}
	return passwordHash, nil
}

// CreateUser inserts a new user with the given username and password hash.
func (ps *PostgresStorage) CreateUser(ctx context.Context, username string, passwordHash string) error {
	_, err := ps.db.Exec(ctx, insertUserTemplate, username, passwordHash)
	if err != nil {
		return err
	}

	return nil
}

func (ps *PostgresStorage) RegisterOrder(ctx context.Context, orderNumber, username string) (model.OrderWriteResult, error) {
	var result model.OrderWriteResult
	var orderID string

	err := ps.db.QueryRow(ctx, insertOrderTemplate, orderNumber, username).Scan(&orderID)
	if err != nil {
		return model.OrderWriteResult{}, err
	}

	err = ps.db.QueryRow(ctx, getOrderTemplate, orderNumber).Scan(&result.Owner)
	if err != nil {
		return model.OrderWriteResult{}, err
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return model.OrderWriteResult{Inserted: true, Owner: username}, nil
	}

	return model.OrderWriteResult{Inserted: true, Owner: username}, nil
}

// withRetries runs operation, retrying it with backoff according to ps.timeouts
// when the classifier marks the error as retriable.
func (ps *PostgresStorage) withRetries(ctx context.Context, operation func() error) error {
	return nil
}

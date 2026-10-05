package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// PostgresStorage implements the service.Repository and accrual.Repository
// interfaces on top of PostgreSQL.
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

// withRetries runs operation, retrying it with backoff according to ps.timeouts
// when the classifier marks the error as retriable.
func (ps *PostgresStorage) withRetries(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt < len(ps.timeouts); attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		classification := ps.classifier.Classify(err)

		if classification == NonRetriable {
			return err
		}
		lastErr = err

		ticker := time.NewTicker(ps.timeouts[attempt])
		select {
		case <-ticker.C:
			ticker.Stop()
		case <-ctx.Done():
			ticker.Stop()
			return ctx.Err()
		}
	}

	// last attempt
	lastErr = operation()
	if lastErr == nil {
		return nil
	}
	classification := ps.classifier.Classify(lastErr)

	if classification == NonRetriable {
		return lastErr
	}

	return fmt.Errorf("operation aborted after %d attempts: %w", len(ps.timeouts)+1, lastErr)
}

//go:generate mockgen -source=common.go -destination=mocks/mock_common.go -package=mocks
package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/services/db/pgerrors"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	RetryingAttempts  = 3
	RetryingDelay     = 100 * time.Millisecond
	RetryingDelayStep = 200 * time.Millisecond

	PsqlGophermartSchema = "gophermart"
)

type (
	txWrapperFunc func(tx *sql.Tx) error
	wrapperFunc   func(ctx context.Context) (*sql.Row, error)

	SQLStorage struct {
		db *sql.DB
	}

	SQLStorageInterface interface {
		AddUser(ctx context.Context, user *models.User) error
		GetUser(ctx context.Context, login string) (*models.User, error)
		GetUserByID(ctx context.Context, userID string) (*models.User, error)

		AddOrder(ctx context.Context, order *models.Order) error
		GetOrder(ctx context.Context, orderID, userID string) (*models.Order, error)
		ListOrders(ctx context.Context, userID string) (models.OrdersList, error)
		UpdateOrder(ctx context.Context, order *models.Order) error

		AddWithdrawn(ctx context.Context, withdrawn *models.Withdrawn) error
		GetWithdrawn(ctx context.Context, userID string) (*models.Withdrawn, error)
		ListWithdrawals(ctx context.Context, userID string) (models.WithdrawnList, error)

		Ping(ctx context.Context) error
	}
)

/*
Конструктор структуры для работы с БД
*/
func NewSQLStorage(psql *sql.DB) SQLStorageInterface {
	return &SQLStorage{db: psql}
}

/*
Пингуем сервер БД
*/
func (r *SQLStorage) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *SQLStorage) withRetry(ctx context.Context, action wrapperFunc) (*sql.Row, error) {
	var lastErr error

	for attempts := range RetryingAttempts {
		obj, err := action(ctx)
		lastErr = err

		if err == nil {
			return obj, nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrors.ClassifyPgError(pgErr) == pgerrors.Retriable {
			delay := RetryingDelay + time.Duration(attempts)*RetryingDelayStep

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				continue
			}
		}

		return obj, err
	}

	return nil, lastErr
}

func (r *SQLStorage) withTxRetry(ctx context.Context, action txWrapperFunc) error {
	var lastErr error

	for attempts := range RetryingAttempts {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		err = action(tx)
		if err == nil {
			return tx.Commit()
		}

		_ = tx.Rollback()
		lastErr = err

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrors.ClassifyPgError(pgErr) == pgerrors.Retriable {
			delay := RetryingDelay + time.Duration(attempts)*RetryingDelayStep

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}

		return err
	}

	return lastErr
}

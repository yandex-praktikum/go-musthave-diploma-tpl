package postgresql_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock" // Стандартная библиотека для моков SQL
	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/repositories/postgresql"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLStorage_Ping(t *testing.T) {
	t.Run("successful ping", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing()

		storage := postgresql.NewSQLStorage(db)
		err = storage.Ping(context.Background())

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ping error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing().WillReturnError(errors.New("connection refused"))

		storage := postgresql.NewSQLStorage(db)
		err = storage.Ping(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection refused")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// тестируем приватный метод withTxRetry через AddUser
func TestSQLStorage_withTxRetry_via_AddUser(t *testing.T) {
	t.Run("successful add user", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		storage := postgresql.NewSQLStorage(db)

		// Ожидаем начало транзакции
		mock.ExpectBegin()
		// Ожидаем Exec внутри транзакции.
		mock.ExpectQuery("INSERT INTO ").
			WithArgs("testuser", sqlmock.AnyArg()). // Логин и любой хеш
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		// Ожидаем коммит
		mock.ExpectCommit()

		user := &models.User{
			Login: "testuser",
			Hash:  "hash",
		}

		err = storage.AddUser(context.Background(), user)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("retry on connection error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		storage := postgresql.NewSQLStorage(db)

		// Симулируем ошибку, которую можно повторить (Code 08006 - Connection Failure)
		pgErr := &pgconn.PgError{
			Code:    pgerrcode.ConnectionFailure,
			Message: "connection error",
		}

		for i := 0; i < postgresql.RetryingAttempts; i++ {
			mock.ExpectBegin()
			mock.ExpectQuery("INSERT").
				WillReturnError(pgErr)
			mock.ExpectRollback()
		}

		user := &models.User{Login: "test", Hash: "hash"}
		err = storage.AddUser(context.Background(), user)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

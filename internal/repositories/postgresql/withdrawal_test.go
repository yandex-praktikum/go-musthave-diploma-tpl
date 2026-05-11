package postgresql_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/repositories/postgresql"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var UserID = "some-user-id"

func TestSQLStorage_GetWithdrawn(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := postgresql.NewSQLStorage(db)

	t.Run("successful get withdrawn", func(t *testing.T) {
		userID := UserID
		// QuerySelectWithdrawn: SELECT user_id, SUM(withdrawn) ...
		rows := sqlmock.NewRows([]string{"user_id", "sum"}).
			AddRow(userID, 500)

		mock.ExpectQuery("SELECT").
			WithArgs(userID).
			WillReturnRows(rows)

		wd, err := storage.GetWithdrawn(context.Background(), userID)
		require.NoError(t, err)
		assert.Equal(t, userID, wd.UserID)
		assert.Equal(t, models.Accrual(500), wd.Sum)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("withdrawn not found", func(t *testing.T) {
		userID := "user_2"

		mock.ExpectQuery("SELECT").
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		wd, err := storage.GetWithdrawn(context.Background(), userID)
		require.Error(t, err)
		assert.Nil(t, wd)
		assert.ErrorIs(t, err, models.ErrWithdrawnNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("retry on connection error then success", func(t *testing.T) {
		userID := "user_retry"
		pgErr := &pgconn.PgError{Code: "08006"} // Connection failure

		// Первая попытка - ошибка
		mock.ExpectQuery("SELECT").
			WithArgs(userID).
			WillReturnError(pgErr)

		// Вторая попытка - успех
		rows := sqlmock.NewRows([]string{"user_id", "sum"}).
			AddRow(userID, 100)
		mock.ExpectQuery("SELECT").
			WithArgs(userID).
			WillReturnRows(rows)

		wd, err := storage.GetWithdrawn(context.Background(), userID)
		require.NoError(t, err)
		assert.Equal(t, models.Accrual(100), wd.Sum)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLStorage_ListWithdrawals(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := postgresql.NewSQLStorage(db)

	t.Run("successful list", func(t *testing.T) {
		userID := UserID
		now := time.Now().Truncate(time.Microsecond)

		// QuerySelectWithdrawnList: SELECT id, user_id, order_id, withdrawn, processed_at
		rows := sqlmock.NewRows([]string{"id", "user_id", "order_id", "withdrawn", "processed_at"}).
			AddRow(1, userID, "order1", 100, now).
			AddRow(2, userID, "order2", 200, now)

		mock.ExpectQuery("SELECT").
			WithArgs(userID).
			WillReturnRows(rows)

		list, err := storage.ListWithdrawals(context.Background(), userID)
		require.NoError(t, err)
		require.Len(t, list, 2)
		assert.Equal(t, "order1", list[0].OrderID)
		assert.Equal(t, models.Accrual(100), list[0].Sum)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty list", func(t *testing.T) {
		userID := "user_empty"

		rows := sqlmock.NewRows([]string{"id", "user_id", "order_id", "withdrawn", "processed_at"})

		mock.ExpectQuery("SELECT").
			WithArgs(userID).
			WillReturnRows(rows)

		list, err := storage.ListWithdrawals(context.Background(), userID)
		require.NoError(t, err)
		assert.Empty(t, list)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		userID := "user_err"

		mock.ExpectQuery("SELECT").
			WithArgs(userID).
			WillReturnError(errors.New("query failed"))

		list, err := storage.ListWithdrawals(context.Background(), userID)
		require.Error(t, err)
		assert.Nil(t, list)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLStorage_AddWithdrawn(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := postgresql.NewSQLStorage(db)

	t.Run("successful add withdrawn", func(t *testing.T) {
		wd := &models.Withdrawal{
			UserID:  UserID,
			OrderID: "order_1",
			Sum:     500,
		}

		mock.ExpectBegin()

		// Insert into withdrawns
		mock.ExpectQuery("INSERT").
			WithArgs(wd.UserID, wd.OrderID, wd.Sum).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Update user balance (decrease)
		mock.ExpectQuery("UPDATE").
			WithArgs(wd.UserID, -wd.Sum.Int()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(wd.UserID))

		mock.ExpectCommit()

		err := storage.AddWithdrawn(context.Background(), wd)
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error during insert", func(t *testing.T) {
		wd := &models.Withdrawal{
			UserID:  UserID,
			OrderID: "order_err",
			Sum:     100,
		}

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT").
			WithArgs(wd.UserID, wd.OrderID, wd.Sum).
			WillReturnError(errors.New("insert failed"))
		mock.ExpectRollback()

		err := storage.AddWithdrawn(context.Background(), wd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insert failed")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error during balance update", func(t *testing.T) {
		wd := &models.Withdrawal{
			UserID:  UserID,
			OrderID: "order_balance_err",
			Sum:     100,
		}

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT").
			WithArgs(wd.UserID, wd.OrderID, wd.Sum).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Ошибка при обновлении баланса (например, недостаточно средств)
		mock.ExpectQuery("UPDATE").
			WithArgs(wd.UserID, -wd.Sum.Int()).
			WillReturnError(errors.New("insufficient funds"))
		mock.ExpectRollback()

		err := storage.AddWithdrawn(context.Background(), wd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("retry logic on transaction", func(t *testing.T) {
		wd := &models.Withdrawal{
			UserID:  "user_retry",
			OrderID: "order_retry",
			Sum:     100,
		}
		pgErr := &pgconn.PgError{Code: pgerrcode.ConnectionFailure}

		// Первая попытка - ошибка соединения при INSERT
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT").
			WithArgs(wd.UserID, wd.OrderID, wd.Sum).
			WillReturnError(pgErr)
		mock.ExpectRollback()

		// Вторая попытка - успех
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT").
			WithArgs(wd.UserID, wd.OrderID, wd.Sum).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
		mock.ExpectQuery("UPDATE").
			WithArgs(wd.UserID, -wd.Sum.Int()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(wd.UserID))
		mock.ExpectCommit()

		err := storage.AddWithdrawn(context.Background(), wd)
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

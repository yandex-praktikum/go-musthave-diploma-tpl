package postgresql_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock" // Стандартная библиотека для моков SQL
	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/repositories/postgresql"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// тест для AddUser
func TestSQLStorage_AddUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := postgresql.NewSQLStorage(db)

	t.Run("successful add user", func(t *testing.T) {
		user := &models.User{Login: "newuser", Hash: "secret_hash"}

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT").
			WithArgs(user.Login, user.Hash).
			// Мы возвращаем одну строку с колонкой "id" и значением 1
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		err := storage.AddUser(context.Background(), user)
		require.NoError(t, err)
		// Проверяем, что ID записался в структуру пользователя
		assert.Equal(t, "1", user.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user already exists (unique violation)", func(t *testing.T) {
		user := &models.User{Login: "existinguser", Hash: "secret"}

		pgErr := &pgconn.PgError{Code: pgerrcode.UniqueViolation, Message: "duplicate key value"}

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT").
			WithArgs(user.Login, user.Hash).
			WillReturnError(pgErr)
		mock.ExpectRollback()

		err := storage.AddUser(context.Background(), user)
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrUserExists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database internal error", func(t *testing.T) {
		user := &models.User{Login: "dberror", Hash: "secret"}

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT").
			WithArgs(user.Login, user.Hash).
			WillReturnError(errors.New("some internal error"))
		mock.ExpectRollback()

		err := storage.AddUser(context.Background(), user)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "some internal error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("retry logic on insert", func(t *testing.T) {
		user := &models.User{Login: "retry_insert", Hash: "hash"}
		pgErr := &pgconn.PgError{Code: pgerrcode.ConnectionFailure}

		// Два раза возвращаем ошибку
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT").WithArgs(user.Login, user.Hash).WillReturnError(pgErr)
		mock.ExpectRollback()

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT").WithArgs(user.Login, user.Hash).WillReturnError(pgErr)
		mock.ExpectRollback()

		// На третий раз успех
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT").
			WithArgs(user.Login, user.Hash).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		err := storage.AddUser(context.Background(), user)
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLStorage_GetUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := postgresql.NewSQLStorage(db)

	t.Run("successful get user", func(t *testing.T) {
		login := "testuser"
		createdAt := time.Now().Truncate(time.Microsecond) // Truncate для сравнения

		rows := sqlmock.NewRows([]string{"id", "login", "hash", "version", "created_at"}).
			AddRow(1, login, "hashed_password", 1, createdAt)

		mock.ExpectQuery("SELECT").
			WithArgs(login).
			WillReturnRows(rows)

		user, err := storage.GetUser(context.Background(), login)
		require.NoError(t, err)
		assert.Equal(t, login, user.Login)
		assert.Equal(t, "hashed_password", user.Hash)
		assert.Equal(t, 1, user.Version)
		assert.Equal(t, createdAt, user.CreatedAt)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		login := "nonexistent"

		// sql.ErrNoRows возвращается при Scan, если запрос вернул 0 строк.
		// sqlmock эмулирует это через WillReturnError или пустые rows + ошибка скана.
		// Проще всего сразу вернуть ошибку.
		mock.ExpectQuery("SELECT").
			WithArgs(login).
			WillReturnError(sql.ErrNoRows)

		user, err := storage.GetUser(context.Background(), login)
		require.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("retry on connection error then success", func(t *testing.T) {
		login := "retryuser"
		pgErr := &pgconn.PgError{Code: "08006"} // Connection failure (Retriable)

		// Первая попытка - ошибка соединения
		mock.ExpectQuery("SELECT").
			WithArgs(login).
			WillReturnError(pgErr)

		// Вторая попытка - успех
		rows := sqlmock.NewRows([]string{"id", "login", "hash", "version", "created_at"}).
			AddRow(2, login, "hash_retry", 1, time.Now())
		mock.ExpectQuery("SELECT").
			WithArgs(login).
			WillReturnRows(rows)

		user, err := storage.GetUser(context.Background(), login)
		require.NoError(t, err)
		assert.Equal(t, "hash_retry", user.Hash)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

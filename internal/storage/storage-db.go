package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	customerror "github.com/with0p/gophermart/internal/custom-error"
)

type StorageDB struct {
	db *sql.DB
}

func NewStorageDB(ctx context.Context, db *sql.DB) (*StorageDB, error) {
	err := initTable(ctx, db)
	if err != nil {
		return nil, err
	}
	return &StorageDB{db: db}, nil
}

func initTable(ctx context.Context, db *sql.DB) error {
	tr, errTr := db.BeginTx(ctx, nil)
	if errTr != nil {
		return errTr
	}

	query := `
    CREATE TABLE IF NOT EXISTS user_auth (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        login TEXT NOT NULL,
        password TEXT NOT NULL
    );`

	tr.ExecContext(ctx, query)

	tr.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS user_login_index ON user_auth (login)`)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return tr.Commit()
	}
}

func (s *StorageDB) CreateUser(ctx context.Context, login string, password string) error {
	query := `
		INSERT INTO user_auth (login, password)
		VALUES ($1, $2)`

	_, errInsert := s.db.ExecContext(ctx, query, login, password)

	if errInsert != nil {
		var pgErr *pgconn.PgError
		if errors.As(errInsert, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			errInsert = customerror.ErrUniqueKeyConstrantViolation
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return errInsert
	}
}

func (s *StorageDB) GetUserID(ctx context.Context, login string, password string) (string, error) {
	query := `SELECT id::text FROM user_auth WHERE login = $1 AND password = $2`

	var userID string
	err := s.db.QueryRowContext(ctx, query, login, password).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", customerror.ErrNoSuchUser
		}
		return "", err
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		return userID, nil
	}
}

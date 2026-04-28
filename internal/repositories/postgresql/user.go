package postgresql

import (
	"context"
	"database/sql"
	"errors"

	model "github.com/KalessinD/gophermart/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	PsqlUserTable = `"` + PsqlGophermartSchema + `"."users"`

	QueryInsertUser = `INSERT INTO ` + PsqlUserTable + ` AS t (login, hash) VALUES($1, $2) RETURNING id`

	QuerySelectUser = `SELECT id, login, hash, version, created_at FROM ` + PsqlUserTable + ` WHERE login = $1`
)

func (r *SQLStorage) GetUser(ctx context.Context, login string) (*model.User, error) {
	user := &model.User{}

	_, err := r.withRetry(ctx, func(ctx context.Context) (*sql.Row, error) {
		row := r.db.QueryRowContext(ctx, QuerySelectUser, login)
		if err := row.Scan(&user.ID, &user.Login, &user.Hash, &user.Version, &user.CreatedAt); err != nil {
			return nil, err
		}
		return row, nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *SQLStorage) AddUser(ctx context.Context, user *model.User) error {
	return r.withTxRetry(ctx, func(tx *sql.Tx) error {
		if err := r.addUserTx(ctx, tx, user); err != nil {
			return err
		}
		return nil
	})
}

func (r *SQLStorage) addUserTx(ctx context.Context, tx *sql.Tx, user *model.User) error {
	err := tx.QueryRowContext(ctx, QueryInsertUser, user.Login, user.Hash).Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return model.ErrUserExists
			}
		}
		return err
	}
	return nil
}

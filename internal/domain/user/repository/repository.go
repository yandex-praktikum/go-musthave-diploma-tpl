package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/benderr/gophermart/internal/domain/user"
	"github.com/benderr/gophermart/internal/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type userRepository struct {
	db  *sql.DB
	log logger.Logger
}

func New(db *sql.DB, log logger.Logger) *userRepository {
	return &userRepository{db: db, log: log}
}

func (u *userRepository) GetUserByLogin(ctx context.Context, login string) (*user.User, error) {

	row := u.db.QueryRowContext(ctx, "SELECT id, login, passhash, created_at from users WHERE login = $1", login)
	var usr user.User
	err := row.Scan(&usr.ID, &usr.Login, &usr.Password, &usr.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}

		return nil, err
	}

	return &usr, nil
}

func (u *userRepository) AddUser(ctx context.Context, login, passhash string) (*user.User, error) {
	_, err := u.db.ExecContext(ctx, `INSERT INTO users (login, passhash) VALUES($1, $2)`, login, passhash)
	if err != nil {
		var perr *pgconn.PgError
		if errors.As(err, &perr) && perr.Code == pgerrcode.UniqueViolation {
			return nil, user.ErrLoginExist
		}

		return nil, err
	}

	return u.GetUserByLogin(ctx, login)
}

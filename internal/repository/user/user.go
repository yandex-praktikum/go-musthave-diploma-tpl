package sql

import (
	"context"
	"database/sql"
	"errors"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type storage struct {
	logger logging.Logger
	db     DB
}

type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Ping(ctx context.Context) error
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

var ErrNotUniqueLogin = errors.New("пользователь с таким логином уже зарегистрирован")

func NewStorage(db DB, logger logging.Logger) *storage {
	return &storage{
		db:     db,
		logger: logger,
	}
}

func (s *storage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (u *storage) Register(ctx context.Context, userRegister *entity.UserRegisterJSON) (*entity.UserDB, error) {
	var userDB entity.UserDB
	err := u.db.QueryRow(
		ctx,
		"INSERT INTO gophermart.users (login, password) values ($1, $2) RETURNING id, login, password",
		userRegister.Login,
		userRegister.Password,
	).Scan(&userDB.ID, &userDB.Login, &userDB.Password)

	// Проверка на то что логин не уникальный
	if err != nil {
		if pgError := err.(*pgconn.PgError); errors.Is(err, pgError) {
			return nil, ErrNotUniqueLogin
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotUniqueLogin
	}
	if err != nil {
		u.logger.Error(err)
		return nil, err
	}

	return &userDB, nil
}

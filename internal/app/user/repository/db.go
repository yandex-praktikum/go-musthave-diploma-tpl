package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/korol8484/gofermart/internal/app/domain"
)

type DbStore struct {
	db *sql.DB
}

func NewDbStore(db *sql.DB) *DbStore {
	return &DbStore{db: db}
}

func (d *DbStore) AddUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	var id domain.UserId

	err := d.db.QueryRowContext(
		ctx,
		`INSERT INTO "user" (login, password_hash) VALUES ($1, $2) returning id;`,
		user.Login, user.PasswordHash,
	).Scan(&id)
	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) {
			if e.Code == "23505" {
				return nil, domain.ErrIssetUser
			}
		}

		return nil, err
	}

	user.ID = id

	return user, nil
}

func (d *DbStore) FindByLogin(ctx context.Context, login string) (*domain.User, error) {
	u := &domain.User{}

	err := d.db.QueryRowContext(ctx, `SELECT u.id, u.login, u.password_hash FROM "user" u WHERE u.login = $1;`, login).Scan(
		&u.ID,
		&u.Login,
		&u.PasswordHash,
	)
	if err != nil {
		return nil, err
	}

	return u, nil
}

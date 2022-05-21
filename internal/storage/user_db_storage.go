package storage

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

const (
	StateActive   = "active"
	StateDisabled = "disabled"

	CreateStateEnum = `create type state as enum ('active', 'disabled');`

	CreateUsersTableScheme = `
       create table users (
			id bigserial PRIMARY KEY,
			name varchar(8192) not null UNIQUE,
			secret bytea not null,
			added timestamptz not null DEFAULT now(),
			flags state not null DEFAULT 'active'
		);`

	CreateUserNameIndex = `create index username_idx on users(name);`

	CheckUsersTable = `select count(*) from users;`

	AddUserQuery = `insert into users (name, secret) values ($1, $2);`

	DatabaseOperationTimeout = 5 * time.Second

	UniqueViolationCode = "23505"
)

type pgxStorage struct {
	ctx    context.Context
	dbConn *pgxpool.Pool
}

func NewUserDatabaseStorage(ctx context.Context, connection *pgxpool.Pool) (UserStorage, error) {
	if err := connection.Ping(ctx); err != nil {
		return nil, err
	}

	if err := prepareDatabase(ctx, connection); err != nil {
		return nil, err
	}

	storage := &pgxStorage{
		ctx:    ctx,
		dbConn: connection,
	}
	return storage, nil
}

func (p *pgxStorage) Add(auth *UserAuthorization) error {
	opCtx, cancel := context.WithTimeout(p.ctx, DatabaseOperationTimeout)
	defer cancel()

	tx, err := p.dbConn.Begin(opCtx)
	if err != nil {
		return err
	}
	defer tx.Rollback(p.ctx)

	_, err = tx.Exec(opCtx, AddUserQuery, auth.UserName, auth.Secret)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == UniqueViolationCode {
				return ErrDuplicateUser
			}
		}
		return err
	}

	return tx.Commit(opCtx)
}

func (p *pgxStorage) Get(userName string) (UserAuthorization, error) {
	return UserAuthorization{}, nil
}

func prepareDatabase(ctx context.Context, conn *pgxpool.Pool) error {
	if r, err := conn.Query(ctx, CheckUsersTable); err == nil {
		r.Close()
		return r.Err()
	}

	r, err := conn.Query(ctx, CreateStateEnum)
	if err != nil {
		return err
	}
	if r.Err() != nil {
		return err
	}
	r.Close()

	r, err = conn.Query(ctx, CreateUsersTableScheme)
	if err != nil {
		return err
	}
	if r.Err() != nil {
		return err
	}
	r.Close()

	r, err = conn.Query(ctx, CreateUserNameIndex)
	if err != nil {
		return err
	}
	r.Close()

	return r.Err()
}

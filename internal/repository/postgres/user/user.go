package user

import (
	"context"
	"database/sql"
)

type UserRepository interface {
	Set(ctx context.Context, login, passwordHash string) error
	GetPasswordHash(ctx context.Context, login string) (string, error)
}

type UserStorageDB struct {
	db *sql.DB
}

func New(db *sql.DB) UserRepository {
	initDB(db)
	return &UserStorageDB{db: db}
}

func (u *UserStorageDB) Set(ctx context.Context, login, passwordHash string) error {
	_, err := u.db.ExecContext(ctx,
		`INSERT INTO users
		(login, password_hash)
		VALUES($1, $2)`,
		login, passwordHash)
	return err
}

func (u *UserStorageDB) GetPasswordHash(ctx context.Context, login string) (string, error) {
	var passwordHash string
	row := u.db.QueryRowContext(ctx,
		`SELECT password_hash
		FROM users
		WHERE login = $1`, login)

	err := row.Scan(&passwordHash)

	return passwordHash, err
}

func initDB(db *sql.DB) {
	db.ExecContext(context.Background(),
		`CREATE TABLE IF NOT EXIST users(
		login VARCHAR PRIMARY KEY
		password_hash VARCHAR
	)`)
}

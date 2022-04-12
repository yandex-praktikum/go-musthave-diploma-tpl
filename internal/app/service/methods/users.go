package methods

import (
	"context"
	"errors"
	db "github.com/EestiChameleon/BonusServiceGOphermart/internal/app/storage"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"time"
)

var (
	ctx                 = context.Background()
	ErrLoginUnavailable = errors.New("provided login is unavailable")
)

type User struct {
	ID           int       `json:"id"`
	RegisteredAt time.Time `json:"registered_at"`
	Login        string    `json:"login"`
	Password     string    `json:"password"`
	LastSignIn   time.Time `json:"last_sign_in"`
}

func NewUser(login, pass string) *User {
	return &User{
		ID:           0,
		RegisteredAt: time.Now(),
		Login:        login,
		Password:     pass,
		LastSignIn:   time.Now(),
	}
}

func (u *User) GetByLogin() error {
	err := pgxscan.Get(context.Background(), db.Pool.DB, u,
		"SELECT id, registered_at, password, last_sign_in FROM users WHERE login=$1", u.Login)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.ErrNotFound
		}
		return err
	}

	return nil
}

func (u *User) Add() error {
	err := db.Pool.DB.QueryRow(ctx,
		"INSERT INTO users(registered_at, login, password, last_sign_in) "+
			"VALUES ($1, $2, $3, $4) "+
			"ON CONFLICT DO NOTHING "+
			"RETURNING id;",
		u.RegisteredAt, u.Login, u.Password, u.LastSignIn).Scan(&u.ID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrLoginUnavailable
		}
		return err
	}

	return nil
}

package service

import (
	"context"
	"database/sql"
	"time"
)

type Storage interface {
	Save(query string, args ...interface{}) error
	SaveTableUser(login, passwordHash string) error
	SaveTableUserAndUpdateToken(login, accessToken string) error
	Get(query string, args ...interface{}) (*sql.Row, error)
	GetUserByAccessToken(order string, login string, now time.Time) error
	CheckTableUserLogin(ctx context.Context, login string) error
	CheckTableUserPassword(ctx context.Context, password string) (string, bool)
}

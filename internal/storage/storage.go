package storage

import "context"

type Storage interface {
	CreateUser(ctx context.Context, login string, password string) error
	GetUserID(ctx context.Context, login string, password string) (string, error)
}

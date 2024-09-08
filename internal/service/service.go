package service

import "context"

type Service interface {
	RegisterUser(ctx context.Context, login string, password string) error
	AuthenticateUser(ctx context.Context, login string, password string) error
}

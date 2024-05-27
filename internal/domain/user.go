package domain

import "context"

type UserRepository interface {
	Insert(ctx context.Context, login, password string) (int, error)
	GetByID(ctx context.Context, id int) (User, error)
}

type User struct {
	ID           int    `json:"id"`
	Login        string `json:"login"`
	PasswordHash string `json:"-"`
}

package domain

import "errors"

type UserId int64

type User struct {
	ID           UserId
	Login        string
	PasswordHash string
}

var (
	ErrIssetUser = errors.New("логин уже занят")
	ErrNotFound  = errors.New("user not found")
)

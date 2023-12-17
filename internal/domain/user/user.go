package user

import (
	"errors"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Login     string    `json:"login" validate:"required"`
	Password  string    `json:"password" validate:"required"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	ErrBadPass    = errors.New("bad pass")
	ErrNotFound   = errors.New("user not found")
	ErrLoginExist = errors.New("login already exist")
)

//Валидация

// type UserUsecase interface {
// 	Register(ctx context.Context, login string, password string) error
// 	GetByLogin(ctx context.Context, login string, password string) (User, error)
// }

// type UserRepository interface {
// 	Get(ctx context.Context, login string, password string) (User, error)
// 	Insert(ctx context.Context, u *User) error
// }

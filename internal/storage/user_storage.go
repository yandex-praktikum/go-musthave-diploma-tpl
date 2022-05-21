package storage

import "errors"

var (
	ErrDuplicateUser = errors.New("duplicate user")
)

type UserAuthorization struct {
	UserName string
	Secret   []byte
}

type UserStorage interface {
	Add(auth *UserAuthorization) error
	Get(userName string) (UserAuthorization, error)
}

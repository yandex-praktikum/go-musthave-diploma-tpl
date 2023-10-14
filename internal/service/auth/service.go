package auth

import (
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/models"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/repository/user/auth"
)

type Service struct {
	auth   auth.Repository
	secret string
}

type Auth interface {
	Registration(models.User) (string, error)
	Login(models.User) (string, error)
	FindUserByID(string) bool
}

func NewService(auth auth.Repository, key string) Auth {
	return &Service{
		auth:   auth,
		secret: key,
	}
}

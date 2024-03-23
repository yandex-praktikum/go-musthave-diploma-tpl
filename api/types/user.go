package types

import (
	"github.com/A-Kuklin/gophermart/internal/domain/entities"
	"github.com/A-Kuklin/gophermart/internal/usecases"
)

type UserCreateRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserCreateResponse struct {
	User  *entities.User `json:"user"`
	Token string         `json:"token"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (c *UserCreateRequest) ToDomain() usecases.UserCreateArgs {
	return usecases.UserCreateArgs{
		Login:    c.Login,
		Password: []byte(c.Password),
	}
}

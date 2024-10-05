package service

import (
	"context"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal/domain"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal/infra"
)

type UserService struct {
	userRepo domain.UserRepository
}

func NewUserService(userRepo domain.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (r UserService) RegisterUser(ctx context.Context, login, password string) (int, error) {
	passwordHash, err := infra.HashPassword(password)
	if err != nil {
		return 0, err
	}

	return r.userRepo.Insert(ctx, login, passwordHash)
}

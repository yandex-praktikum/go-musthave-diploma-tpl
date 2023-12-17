package usecase

import (
	"context"

	"github.com/benderr/gophermart/internal/domain/user"
	"github.com/benderr/gophermart/internal/logger"
	"github.com/benderr/gophermart/internal/utils"
)

type userRepo interface {
	GetUserByLogin(ctx context.Context, login string) (*user.User, error)
	AddUser(ctx context.Context, login, passhash string) (*user.User, error)
}

type userUsecase struct {
	repo   userRepo
	logger logger.Logger
}

func New(repo userRepo, logger logger.Logger) *userUsecase {
	return &userUsecase{repo: repo, logger: logger}
}

func (u *userUsecase) Login(ctx context.Context, login, password string) (*user.User, error) {
	usr, err := u.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	if usr == nil {
		return nil, user.ErrNotFound
	}

	if !utils.CheckPasswordHash(password, usr.Password) {
		return nil, user.ErrBadPass
	}

	return usr, nil
}

func (u *userUsecase) Register(ctx context.Context, login, password string) (*user.User, error) {
	passhash, err := utils.HashPassword(password)

	if err != nil {
		return nil, err
	}

	return u.repo.AddUser(ctx, login, passhash)
}

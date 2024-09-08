package service

import (
	"context"

	"github.com/with0p/gophermart/internal/storage"
	"github.com/with0p/gophermart/internal/utils"
)

type ServiceGophermart struct {
	storage storage.Storage
}

func NewServiceGophermart(currentStorage storage.Storage) ServiceGophermart {
	return ServiceGophermart{
		storage: currentStorage,
	}
}

func (s *ServiceGophermart) RegisterUser(ctx context.Context, login string, password string) error {
	return s.storage.CreateUser(ctx, login, utils.HashPassword(password))
}

func (s *ServiceGophermart) AuthenticateUser(ctx context.Context, login string, password string) error {
	_, errDB := s.storage.GetUserID(ctx, login, utils.HashPassword(password))
	if errDB != nil {
		return errDB
	}

	return nil
}

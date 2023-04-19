package service

import (
	"context"

	"github.com/RedWood011/cmd/gophermart/internal/apperrors"
	"github.com/RedWood011/cmd/gophermart/internal/entity"
)

func (s *Service) CreateUser(ctx context.Context, user entity.User) error {
	err := s.storage.SaveUser(ctx, user)
	if err != nil && isUniqueViolationError(err) {
		return apperrors.ErrUserExists
	}
	return err
}

func (s *Service) IdentificationUser(ctx context.Context, user entity.User) error {
	existUser, err := s.storage.GetUser(ctx, user)
	if err != nil {
		return err
	}
	if !existUser.IsEqual(user) {
		err = apperrors.ErrAuth
	}
	return err
}
func (s *Service) GetUserBalance(ctx context.Context, userID string) (entity.UserBalance, error) {
	return s.storage.GetUserBalance(ctx, userID)

}

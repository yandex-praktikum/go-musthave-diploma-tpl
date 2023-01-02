package usecase

import (
	"context"

	"github.com/brisk84/gofemart/domain"
)

func (s *service) Register(ctx context.Context, user domain.User) (string, error) {
	return s.storage.Register(ctx, user)
}

func (s *service) Login(ctx context.Context, user domain.User) (bool, string, error) {
	return s.storage.Login(ctx, user)
}

func (s *service) Auth(ctx context.Context, token string) (*domain.User, error) {
	return s.storage.Auth(ctx, token)
}

func (s *service) UserOrders(ctx context.Context, user domain.User, order int) error {
	return s.storage.UserOrders(ctx, user, order)
}

func (s *service) UserOrdersGet(ctx context.Context, user domain.User) ([]domain.Order, error) {
	return s.storage.UserOrdersGet(ctx, user)
}

func (s *service) UserBalanceWithdraw(ctx context.Context, user domain.User, withdraw domain.Withdraw) error {
	return s.storage.UserBalanceWithdraw(ctx, user, withdraw)
}

func (s *service) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	userID, err := s.storage.CreateUser(ctx, user)
	if err != nil {
		return domain.User{}, err
	}

	return s.storage.GetUser(ctx, userID)
}

func (s *service) GetUser(ctx context.Context, userID int64) (domain.User, error) {
	return s.storage.GetUser(ctx, userID)
}

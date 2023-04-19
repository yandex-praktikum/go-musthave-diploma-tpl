package service

import (
	"context"

	"github.com/RedWood011/cmd/gophermart/internal/entity"
)

func (s *Service) CreateWithdrawal(ctx context.Context, withdraw entity.Withdraw, userID string) error {
	return s.storage.SaveWithdraw(ctx, withdraw, userID)
}

func (s *Service) GetWithdrawals(ctx context.Context, userID string) ([]entity.Withdraw, error) {
	return s.storage.GetWithdrawals(ctx, userID)
}

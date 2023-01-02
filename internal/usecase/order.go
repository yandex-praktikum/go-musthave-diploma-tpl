package usecase

import (
	"context"

	domain "github.com/brisk84/gofemart/domain"
)

func (s *service) CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	orderID, err := s.storage.CreateOrder(ctx, order)
	if err != nil {
		return domain.Order{}, err
	}

	return s.storage.GetOrder(ctx, orderID)
}

func (s *service) GetOrder(ctx context.Context, orderID int64) (domain.Order, error) {
	return s.storage.GetOrder(ctx, orderID)
}

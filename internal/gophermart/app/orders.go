package app

import (
	"context"

	"github.com/StasMerzlyakov/go-musthave-diploma-tpl/internal/gophermart/domain"
)

func NewOrders() *orders {
	// TODO
	return nil
}

type orders struct {
}

func (ord *orders) Upload(ctx context.Context, orderData *domain.OrderData) error {
	// TODO
	return nil
}

func (ord *orders) Orders(ctx context.Context) ([]domain.OrderData, error) {
	// TODO
	return nil, nil
}

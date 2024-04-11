package app

import (
	"context"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

func NewOrders() *orders {
	// TODO
	return nil
}

type orders struct {
}

// загрузка пользователем номера заказа для расчёта
func (ord *orders) Upload(ctx context.Context, number domain.OrderNumber) error {
	// TODO
	return nil
}

// получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
func (ord *orders) Orders(ctx context.Context) ([]domain.OrderData, error) {
	// TODO
	return nil, nil
}

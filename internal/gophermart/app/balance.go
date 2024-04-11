package app

import (
	"context"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

func NewBalance() *balance {
	// TODO
	return nil
}

type balance struct {
}

// получение текущего баланса счёта баллов лояльности пользователя;
func (b *balance) Balance(ctx context.Context) (*domain.Balance, error) {
	// TODO
	return nil, nil
}

// запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
func (b *balance) Withdraw(ctx context.Context, withdraw *domain.WithdrawData) error {
	// TODO
	return nil
}

// получение информации о выводе средств с накопительного счёта пользователем.
func (b *balance) Withdrawals(ctx context.Context) ([]domain.WithdrawalsData, error) {
	// TODO
	return nil, nil
}

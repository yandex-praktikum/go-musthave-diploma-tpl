package app

import (
	"context"

	"github.com/StasMerzlyakov/go-musthave-diploma-tpl/internal/gophermart/domain"
)

func NewBalance() *balance {
	// TODO
	return nil
}

type balance struct {
}

func (b *balance) Get(ctx context.Context) (*domain.Balance, error) {
	// TODO
	return nil, nil
}

func (b *balance) Withdraw(ctx context.Context, withdraw *domain.WithdrawData) error {
	// TODO
	return nil
}

func (b *balance) List(ctx context.Context, withdraw *domain.WithdrawData) ([]domain.WithdrawData, error) {
	// TODO
	return nil, nil
}

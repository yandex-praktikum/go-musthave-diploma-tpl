package usecase

import (
	"context"

	"github.com/benderr/gophermart/internal/domain/withdrawal"
)

type WithdrawRepo interface {
	GetWithdrawsByUser(ctx context.Context, userid string) ([]withdrawal.Withdrawal, error)
}

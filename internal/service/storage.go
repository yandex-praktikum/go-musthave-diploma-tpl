package service

import (
	"context"

	"github.com/nikitaw13/go-musthave-diploma/internal/model"
)

// Repository is the storage interface consumed by LoyaltyService.
type Repository interface {
	UserExists(ctx context.Context, username string) (bool, error)
	GetPasswordHash(ctx context.Context, username string) (string, error)
	CreateUser(ctx context.Context, username string, passwordHash string) error
	RegisterOrder(ctx context.Context, orderNumber string, username string) (model.OrderWriteResult, error)
	ListOrders(ctx context.Context, username string) ([]model.OrderResponse, error)
	GetBalance(ctx context.Context, username string) (model.BalanceResponse, error)
	Withdraw(ctx context.Context, username string, withdrawal model.WithdrawalRequest) error
	ListWithdrawals(ctx context.Context, username string) ([]model.WithdrawalResponse, error)
}

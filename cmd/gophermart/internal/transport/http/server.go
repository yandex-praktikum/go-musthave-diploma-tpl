package http

import (
	"context"

	"github.com/RedWood011/cmd/gophermart/internal/entity"
)

type ServerInterface interface {
	CreateOrder(ctx context.Context, order entity.Order) error
	GetOrders(ctx context.Context, userID string) ([]entity.Order, error)
	GetWithdrawals(ctx context.Context, userID string) ([]entity.Withdraw, error)
	CreateWithdrawal(ctx context.Context, withdraw entity.Withdraw, userID string) error
	CreateUser(ctx context.Context, user entity.User) error
	IdentificationUser(ctx context.Context, user entity.User) error
	GetBalance(ctx context.Context, userID string) (entity.UserBalance, error)
}

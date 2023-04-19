package service

import (
	"context"

	"github.com/RedWood011/cmd/gophermart/internal/entity"
)

type Storage interface {
	SaveOrder(ctx context.Context, order entity.Order) error
	GetOrder(ctx context.Context, orderNum string) (entity.Order, error)
	GetUserOrders(ctx context.Context, userUID string) ([]entity.Order, error)
	GetAllOrders(ctx context.Context) ([]entity.Order, error)
	UpdateOrders(ctx context.Context, orders []entity.Order) error
	SaveWithdraw(ctx context.Context, withdraw entity.Withdraw, userID string) error
	GetWithdrawals(ctx context.Context, userID string) ([]entity.Withdraw, error)
	SaveUser(ctx context.Context, user entity.User) error
	GetUser(ctx context.Context, user entity.User) (entity.User, error)
	GetUserBalance(ctx context.Context, userID string) (entity.UserBalance, error)
	UpdateUsersBalance(ctx context.Context, users []entity.User) error
}

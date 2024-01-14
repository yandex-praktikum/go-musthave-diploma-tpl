package store

import (
	"context"

	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store/models"
)

type Store interface {
	RegisterUser(ctx context.Context, data models.RegisterData) error
	LoginUser(ctx context.Context, data models.LoginData) (string, error)
	AddOrder(ctx context.Context, data models.AddOrderData) error
	GetOwnerForOrder(ctx context.Context, data models.GetOwnerForOrderData) (string, error)
	GetOrders(ctx context.Context, data models.GetOrdersData) (models.GetOrdersDataResult, error)
	GetUserBalance(ctx context.Context, data models.GetUserBalanceData) (models.GetUserBalanceDataResult, error)
	Withdraw(ctx context.Context, data models.WithdrawData) error
	Withdrawals(ctx context.Context, data models.WithdrawalsData) (models.WithdrawsDataResult, error)

	GetNewOrders(ctx context.Context, owner, count int) (models.LockNewOrders, error)
	RestoreNewOrders(ctx context.Context, owner int) (models.LockNewOrders, error)
	ProcessedOrder(ctx context.Context, o models.OrderData) error
}

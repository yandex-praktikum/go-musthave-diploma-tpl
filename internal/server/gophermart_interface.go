package server

import (
	"context"

	"github.com/Raime-34/gophermart.git/internal/dto"
)

//go:generate mockgen -source=gophermart_interface.go -destination=mocks/gophermart.go -package=mocksgophermart

type gophermartInterface interface {
	RegisterUser(context.Context, dto.UserCredential) error
	LoginUser(context.Context, dto.UserCredential) (*dto.UserData, error)
	InsertOrder(ctx context.Context, orderNumber string) error
	GetUserOrders(ctx context.Context) ([]*dto.GetOrdersInfoResp, error)
	GetUserBalance(context.Context) (*dto.BalanceInfo, error)
	ProcessWithdraw(context.Context, dto.WithdrawRequest) error
	GetWithdraws(context.Context) ([]*dto.WithdrawInfo, error)
}

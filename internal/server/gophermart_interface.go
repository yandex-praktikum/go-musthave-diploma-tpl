package server

import (
	"context"

	"github.com/Raime-34/gophermart.git/internal/dto"
)

type gophermartInterface interface {
	RegisterUser(context.Context, dto.UserCredential) error
	LoginUser(context.Context, dto.UserCredential) (*dto.UserData, error)
	InsertOrder(ctx context.Context, orderNumber string) error
	// GetUserOrders() []dto.OrderInfo
	// GetUserBalance() []dto.BalanceInfo
	// ProcessWithdraw(dto.WithdrawRequest) error
	// GetWithdraws() []dto.WithdrawInfo
}

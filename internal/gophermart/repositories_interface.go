package gophermart

import (
	"context"

	"github.com/Raime-34/gophermart.git/internal/dto"
)

//go:generate mockgen -source=repositories_interface.go -destination=../../mocks/repositories.go -package=mock
type repositoriesInt interface {
	UpdateOrder(ctx context.Context, state dto.AccrualCalculatorDTO) error
	RegisterOrder(ctx context.Context, orderNumber string) error
	GetOrders(ctx context.Context) ([]*dto.OrderInfo, error)

	GetUser(ctx context.Context, userInfo dto.UserCredential) (*dto.UserData, error)
	RegisterUser(ctx context.Context, userInfo dto.UserCredential) error

	RegisterWithdraw(ctx context.Context, req dto.WithdrawRequest) error
	GetWithdraws(ctx context.Context) ([]*dto.WithdrawInfo, error)
}

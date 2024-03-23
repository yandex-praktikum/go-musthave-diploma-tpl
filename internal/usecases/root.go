package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/A-Kuklin/gophermart/internal/domain/entities"
	"github.com/A-Kuklin/gophermart/internal/storage"
)

type UserUseCasesImpl interface {
	Create(ctx context.Context, args UserCreateArgs) (*entities.User, error)
	Login(ctx context.Context, args UserCreateArgs) (*entities.User, error)
}

type OrderUseCasesImpl interface {
	Create(ctx context.Context, userID uuid.UUID, strOrder string) (*entities.Order, error)
	GetOrders(ctx context.Context, userID uuid.UUID) ([]entities.Order, error)
	GetAccruals(ctx context.Context, userID uuid.UUID) (int64, error)
}

type WithdrawUseCasesImpl interface {
	Create(ctx context.Context, withdraw *entities.Withdraw) (*entities.Withdraw, error)
	GetSumWithdrawals(ctx context.Context, userID uuid.UUID) (int64, error)
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]entities.Withdraw, error)
}

type UseCases struct {
	User     UserUseCasesImpl
	Order    OrderUseCasesImpl
	Withdraw WithdrawUseCasesImpl
}

func NewUseCases(strg *storage.Storager, logger logrus.FieldLogger) *UseCases {
	return &UseCases{
		User:     NewUserUseCases(strg.Users, logger),
		Order:    NewOrderUseCases(strg.Orders, logger),
		Withdraw: NewWithdrawUseCases(strg.Withdrawals, logger),
	}
}

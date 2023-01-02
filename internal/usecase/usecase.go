package usecase

import (
	"context"

	"github.com/brisk84/gofemart/domain"

	"go.uber.org/zap"
)

//go:generate mockery --name=storage --structname=storageMock --filename=storage_mock.go --inpackage
type storage interface {
	Register(ctx context.Context, user domain.User) (string, error)
	Login(ctx context.Context, user domain.User) (bool, string, error)
	Auth(ctx context.Context, token string) (*domain.User, error)
	UserOrders(ctx context.Context, user domain.User, order int) error
	UserOrdersGet(ctx context.Context, user domain.User) ([]domain.Order, error)
	UserBalanceWithdraw(ctx context.Context, user domain.User, withdraw domain.Withdraw) error

	CreateUser(ctx context.Context, user domain.User) (int64, error)
	GetUser(ctx context.Context, userID int64) (domain.User, error)

	CreateOrder(ctx context.Context, order domain.Order) (int64, error)
	GetOrder(ctx context.Context, orderID int64) (domain.Order, error)
}

type service struct {
	logger  *zap.Logger
	storage storage
}

func New(logger *zap.Logger, storage storage) *service {
	return &service{
		logger:  logger,
		storage: storage,
	}
}

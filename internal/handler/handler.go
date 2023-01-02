package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/brisk84/gofemart/domain"
	"go.uber.org/zap"
)

type Handler struct {
	logger  *zap.Logger
	useCase useCase
}

//go:generate mockery --name=useCase --structname=useCaseMock --filename=usecase_mock.go --inpackage
type useCase interface {
	Register(ctx context.Context, user domain.User) (string, error)
	Login(ctx context.Context, user domain.User) (bool, string, error)
	Auth(ctx context.Context, token string) (*domain.User, error)
	UserOrders(ctx context.Context, user domain.User, order int) error
	UserOrdersGet(ctx context.Context, user domain.User) ([]domain.Order, error)
	UserBalanceWithdraw(ctx context.Context, user domain.User, withdraw domain.Withdraw) error

	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUser(ctx context.Context, userID int64) (domain.User, error)

	CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error)
	GetOrder(ctx context.Context, orderID int64) (domain.Order, error)
}

func New(logger *zap.Logger, useCase useCase) *Handler {
	return &Handler{
		logger:  logger,
		useCase: useCase,
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

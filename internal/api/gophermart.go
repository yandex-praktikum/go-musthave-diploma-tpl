//go:generate oapi-codegen --config=../../api/types.cfg.yaml ../../api/api.yaml
//go:generate oapi-codegen --config=../../api/server.cfg.yaml ../../api/api.yaml

package api

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/SmoothWay/gophermart/internal/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Service interface {
	RegisterUser(ctx context.Context, login, password string) error
	Authenticate(ctx context.Context, login, password string) (string, error)
	UploadOrder(ctx context.Context, userID uuid.UUID, orderNumber string) error
	GetOrders(ctx context.Context, userID uuid.UUID) ([]model.Order, error)
	WithdrawalRequest(ctx context.Context, userID uuid.UUID, orderNumber string, sum float64) error
	GetBalance(ctx context.Context, userID uuid.UUID) (float64, float64, error)
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]model.Withdrawal, error)
}

type Gophermart struct {
	logger  *slog.Logger
	service Service
	secret  []byte
}

var _ StrictServerInterface = (*Gophermart)(nil)

var MessageInternalError = "Internal server error"

func NewGophermart(logger *slog.Logger, service Service, secret []byte) *Gophermart {
	return &Gophermart{
		logger:  logger,
		service: service,
		secret:  secret,
	}
}

func (g *Gophermart) RegisterUser(ctx context.Context, request RegisterUserRequestObject) (RegisterUserResponseObject, error) {
	err := g.service.RegisterUser(ctx, request.Body.Login, request.Body.Password)
	if err != nil {
		g.logger.Info("Register user error", slog.String("error", err.Error()))
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code.Name() == "unique_violation" {
				return RegisterUser409JSONResponse{
					Message: "User already exist",
				}, nil
			}
		}

		return RegisterUser500JSONResponse{
			Message: MessageInternalError,
		}, nil
	}

	token, err := g.service.Authenticate(ctx, request.Body.Login, request.Body.Password)
	if err != nil {
		g.logger.Info("Authentication error", slog.String("error", err.Error()))
		return RegisterUser500JSONResponse{
			Message: MessageInternalError,
		}, nil
	}

	return RegisterUser200Response{
		RegisterUser200ResponseHeaders{
			Authorization: "Bearer " + token,
		},
	}, nil
}

func (g *Gophermart) LoginUser(ctx context.Context, request LoginUserRequestObject) (LoginUserResponseObject, error) {
	token, err := g.service.Authenticate(ctx, request.Body.Login, request.Body.Password)
	if err != nil {
		g.logger.Info("Authentication error", slog.String("error", err.Error()))

		if errors.Is(err, sql.ErrNoRows) {
			return LoginUser401JSONResponse{
				Message: "Unauthorized",
			}, nil
		}

		return LoginUser500JSONResponse{
			Message: MessageInternalError,
		}, nil
	}

	return LoginUser200Response{
		LoginUser200ResponseHeaders{
			Authorization: "Bearer " + token,
		},
	}, nil

}

func (g *Gophermart) GetOrders(ctx context.Context, _ GetOrdersRequestObject) (GetOrdersResponseObject, error) {
	userID := FromContext(ctx)

	orders, err := g.service.GetOrders(ctx, userID)
	if err != nil {
		g.logger.Info("GetOrder error", slog.String("error", err.Error()))
		if errors.Is(err, sql.ErrNoRows) {
			return GetOrders204Response{}, nil
		}

		return GetOrders500JSONResponse{
			Message: MessageInternalError,
		}, nil
	}

	GetOrderResponse := make(GetOrders200JSONResponse, len(orders))

	for i, order := range orders {
		accrual := order.Accrual
		GetOrderResponse[i] = Order{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    &accrual,
			UploadedAt: order.UploadedAt,
		}

	}
	return GetOrderResponse, nil
}

func (g *Gophermart) UploadOrder(ctx context.Context, request UploadOrderRequestObject) (UploadOrderResponseObject, error) {

}
func (g *Gophermart) GetBalance(ctx context.Context, userID uuid.UUID) (float64, float64, error) {

}
func (g *Gophermart) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]model.Withdrawal, error) {

}

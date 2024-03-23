//go:generate oapi-codegen --config=../../api/types.cfg.yaml ../../api/api.yaml
//go:generate oapi-codegen --config=../../api/server.cfg.yaml ../../api/api.yaml

package api

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/SmoothWay/gophermart/internal/model"
	postgresrepo "github.com/SmoothWay/gophermart/internal/repository/postgres"
	"github.com/SmoothWay/gophermart/internal/service"
	"github.com/SmoothWay/gophermart/internal/util"
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

var MessageInternalError = "internal server error"

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
					Message: "user already exist",
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
				Message: "unauthorized",
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
	userID := FromContext(ctx)

	if valid := util.IsValid(*request.Body); !valid {
		return UploadOrder422JSONResponse{
			Message: "invalid order number",
		}, nil
	}

	err := g.service.UploadOrder(ctx, userID, *request.Body)
	if err != nil {
		g.logger.Info("Upload order error", slog.String("error", err.Error()))
		var pgErr *pq.Error

		if errors.As(err, &pgErr) {
			if pgErr.Code.Name() == "unique_violation" {
				return UploadOrder409JSONResponse{
					Message: "order already exists",
				}, nil
			}
		}
		if errors.Is(err, service.ErrOrderAlreadyExistThisUser) {
			return UploadOrder200Response{}, nil
		}

		return UploadOrder500JSONResponse{
			Message: MessageInternalError,
		}, nil
	}

	return UploadOrder202Response{}, nil
}
func (g *Gophermart) GetBalance(ctx context.Context, _ GetBalanceRequestObject) (GetBalanceResponseObject, error) {
	userID := FromContext(ctx)

	balance, withdrawn, err := g.service.GetBalance(ctx, userID)
	if err != nil {
		g.logger.Info("GetBalance error", slog.String("error", err.Error()))
		if errors.Is(err, sql.ErrNoRows) {
			return GetBalance200JSONResponse{}, nil
		}

		return GetBalance500JSONResponse{
			Message: MessageInternalError,
		}, nil
	}

	return GetBalance200JSONResponse{
		Current:   balance,
		Withdrawn: withdrawn,
	}, nil

}

func (g *Gophermart) WithdrawalRequest(ctx context.Context, request WithdrawalRequestRequestObject) (WithdrawalRequestResponseObject, error) {

	userID := FromContext(ctx)

	if valid := util.IsValid(request.Body.Order); !valid {
		return WithdrawalRequest422JSONResponse{
			Message: "invalid order number",
		}, nil
	}

	err := g.service.WithdrawalRequest(ctx, userID, request.Body.Order, request.Body.Sum)
	if err != nil {
		g.logger.Info("WithdrwalRequest error", slog.String("error", err.Error()))
		if errors.Is(err, postgresrepo.ErrNotEnoughFunds) {
			return WithdrawalRequest402JSONResponse{
				Message: postgresrepo.ErrNotEnoughFunds.Error(),
			}, nil
		}

		return WithdrawalRequest500JSONResponse{
			Message: MessageInternalError,
		}, nil
	}

	return WithdrawalRequest200Response{}, nil

}

func (g *Gophermart) GetWithdrawals(ctx context.Context, request GetWithdrawalsRequestObject) (GetWithdrawalsResponseObject, error) {
	userID := FromContext(ctx)

	withdrwals, err := g.service.GetWithdrawals(ctx, userID)
	if err != nil {
		g.logger.Info("GetWithdrawls error", slog.String("error", err.Error()))
		if errors.Is(err, sql.ErrNoRows) {
			return GetWithdrawals204Response{}, nil
		}

		return GetWithdrawals500JSONResponse{
			Message: MessageInternalError,
		}, nil
	}

	ws := make(GetWithdrawals200JSONResponse, len(withdrwals))

	for i, withdrawl := range withdrwals {
		processedAt := withdrawl.ProcessedAt

		ws[i] = Withdrawal{
			Order:       withdrawl.Order,
			Sum:         withdrawl.Sum,
			ProcessedAt: &processedAt,
		}
	}

	return ws, nil
}

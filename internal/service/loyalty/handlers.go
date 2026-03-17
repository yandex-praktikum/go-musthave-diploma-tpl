package loyalty

import (
	"context"
	"errors"
	"gmart/internal/domain"
	"gmart/internal/dto"
	"gmart/internal/model/auth"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
)

// ------------ Balance ------------

type balanceResponse struct {
	Body struct {
		Current   domain.Money `json:"current"`
		Withdrawn domain.Money `json:"withdrawn"`
	}
}

func (l *Loyalty) getBalanceHandler() func(ctx context.Context, in *struct{}) (*balanceResponse, error) {
	return func(ctx context.Context, in *struct{}) (*balanceResponse, error) {
		userID, ok := ctx.Value(auth.UserID).(domain.UserID)
		if !ok {
			return nil, huma.Error401Unauthorized("пользователь не авторизован")
		}

		current, withdrawn, err := l.GetBalance(ctx, userID)
		if err != nil {
			slog.Error("get balance fail", "user_id", userID, "err", err)
			return nil, huma.Error500InternalServerError("внутренняя ошибка сервера")
		}

		resp := &balanceResponse{}
		resp.Body.Current = current
		resp.Body.Withdrawn = withdrawn
		return resp, nil
	}
}

// ------------ Withdraw ------------

type withdrawResponse struct {
	Status int `status:"200"`
}

type withdrawInput struct {
	Body struct {
		Order domain.OrderNumber `json:"order" doc:"Номер заказа"`
		Sum   domain.Money       `json:"sum" doc:"Сумма к списанию"`
	}
}

func (l *Loyalty) withdrawHandler() func(ctx context.Context, in *withdrawInput) (*withdrawResponse, error) {
	return func(ctx context.Context, in *withdrawInput) (*withdrawResponse, error) {
		userID, ok := ctx.Value(auth.UserID).(domain.UserID)
		if !ok {
			return nil, huma.Error401Unauthorized("пользователь не авторизован")
		}

		err := l.Withdraw(ctx, userID, in.Body.Order, in.Body.Sum)
		if err != nil {
			if errors.Is(err, ErrInvalidOrderNumber) {
				return nil, huma.Error422UnprocessableEntity("неверный формат номера заказа")
			}

			// Исправлено на Error402PaymentRequired
			if errors.Is(err, ErrInsufficientFunds) {
				return nil, huma.Error402PaymentRequired("на счету недостаточно средств")
			}

			if errors.Is(err, ErrWithdrawConflict) {
				return nil, huma.Error409Conflict("запрос на списание уже зарегистрирован")
			}

			slog.Error("withdraw fail", "user_id", userID, "order", in.Body.Order, "err", err)
			return nil, huma.Error500InternalServerError("внутренняя ошибка сервера")
		}

		return &withdrawResponse{Status: 200}, nil
	}
}

// ------------ Withdrawals List ------------

type withdrawalsResponse struct {
	Status int                  `status:"200"`
	Body   []dto.WithdrawalItem `json:",omitempty"` // omitempty важен для 204
}

func (l *Loyalty) getWithdrawalsHandler() func(ctx context.Context, in *struct{}) (*withdrawalsResponse, error) {
	return func(ctx context.Context, in *struct{}) (*withdrawalsResponse, error) {
		userID, ok := ctx.Value(auth.UserID).(domain.UserID)
		if !ok {
			return nil, huma.Error401Unauthorized("пользователь не авторизован")
		}

		res, err := l.GetWithdrawals(ctx, userID)
		if err != nil {
			slog.Error("get withdrawals fail", "user_id", userID, "err", err)
			return nil, huma.Error500InternalServerError("внутренняя ошибка сервера")
		}

		if len(res) == 0 {
			// Возвращаем 204 без тела
			return &withdrawalsResponse{Status: 204}, nil
		}

		return &withdrawalsResponse{
			Status: 200,
			Body:   res,
		}, nil
	}
}

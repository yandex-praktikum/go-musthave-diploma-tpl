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

type balanceInput struct {
	dto.AuthInput
}

type balanceResponse struct {
	Body struct {
		Current   domain.Amount `json:"current" doc:"Денежная сумма текущего баланса"`
		Withdrawn domain.Amount `json:"withdrawn" doc:"Списанная за все время денежная сумма"`
	}
}

func (l *Loyalty) getBalanceHandler() func(ctx context.Context, in *balanceInput) (*balanceResponse, error) {
	return func(ctx context.Context, in *balanceInput) (*balanceResponse, error) {
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

type withdrawInput struct {
	dto.AuthInput
	Body struct {
		Order domain.OrderNumber `json:"order" doc:"Номер заказа"`
		Sum   domain.Amount      `json:"sum" doc:"Денежная сумма к списанию"`
	}
}

type withdrawResponse struct {
	Status int `status:"200"`
	Body   struct {
		Code   int    `json:"code" doc:"Статус код операции"`
		Status string `json:"status" doc:"Статус операции"`
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

		res := &withdrawResponse{Status: 200}
		res.Body.Code = 200
		res.Body.Status = "Success"

		return res, nil
	}
}

// ------------ Withdrawals List ------------

type withdrawalsInput struct {
	dto.AuthInput
}

type withdrawalsResponse struct {
	Status int                  `status:"200"`
	Body   []dto.WithdrawalItem `json:",omitempty"` // omitempty важен для 204
}

func (l *Loyalty) getWithdrawalsHandler() func(ctx context.Context, in *withdrawalsInput) (*withdrawalsResponse, error) {
	return func(ctx context.Context, in *withdrawalsInput) (*withdrawalsResponse, error) {
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

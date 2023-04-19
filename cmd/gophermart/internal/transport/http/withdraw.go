package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/RedWood011/cmd/gophermart/internal/apperrors"
	"github.com/RedWood011/cmd/gophermart/internal/entity"
	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/gofiber/fiber/v2"
)

type WithdrawResponse struct {
	OrderNumber string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type WithdrawRequest struct {
	OrderNumber string  `json:"order"`
	Sum         float32 `json:"sum"`
}

func (c *Controller) CreateWithdraw() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var withdrawRequest WithdrawRequest

		if err := ctx.BodyParser(&withdrawRequest); err != nil {
			ctx.Status(http.StatusBadRequest)
			return ctx.JSON(ErrorResponse(err))
		}

		err := goluhn.Validate(withdrawRequest.OrderNumber)
		if err != nil {
			ctx.Status(http.StatusUnprocessableEntity)
		}
		err = c.service.CreateWithdrawal(ctx.Context(), entity.Withdraw{
			UserID:      ctx.Get("userID"),
			OrderNumber: withdrawRequest.OrderNumber,
			Sum:         withdrawRequest.Sum,
			ProcessedAt: time.Now(),
		}, ctx.Get("userID"))
		if err != nil {
			switch {
			case errors.Is(err, apperrors.ErrNoMoney):
				ctx.Status(http.StatusPaymentRequired)
				return ctx.JSON(ErrorResponse(err))
			default:
				ctx.Status(http.StatusInternalServerError)
				return ctx.JSON(ErrorResponse(err))
			}
		}
		ctx.Status(http.StatusOK)
		return nil
	}
}

func (c *Controller) GetWithdraws() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		withdrawals, err := c.service.GetWithdrawals(ctx.Context(), ctx.Get("userID"))
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return ctx.JSON(ErrorResponse(err))
		}

		if len(withdrawals) == 0 {
			ctx.Status(http.StatusNoContent)
			return ctx.JSON(ErrorResponse(errors.New("no withdraws")))

		}

		return ctx.JSON(withdrawsFromService(withdrawals))
	}
}

func withdrawsFromService(withdraws []entity.Withdraw) []WithdrawResponse {
	var response []WithdrawResponse
	for _, withdraw := range withdraws {
		response = append(response, WithdrawResponse{
			OrderNumber: withdraw.OrderNumber,
			Sum:         withdraw.Sum,
			ProcessedAt: withdraw.ProcessedAt,
		})
	}
	return response
}

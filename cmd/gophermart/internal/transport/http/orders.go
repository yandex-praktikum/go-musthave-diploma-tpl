package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/RedWood011/cmd/gophermart/internal/apperrors"
	"github.com/RedWood011/cmd/gophermart/internal/entity"
	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type OrderResponse struct {
	Number     string  `json:"number"`
	UploadedAt string  `json:"uploaded_at"`
	Status     string  `json:"status"`
	Accrual    float32 `json:"accrual"`
}

func (c *Controller) CreateOrder() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		body := ctx.Body()
		if len(body) == 0 {
			ctx.Status(http.StatusBadRequest)
			return ctx.JSON(ErrorResponse(errors.New("empty body")))
		}

		err := goluhn.Validate(string(body))
		if err != nil {
			ctx.Status(http.StatusUnprocessableEntity)
			return ctx.JSON(ErrorResponse(err))
		}
		orderNumber := string(body)
		order := entity.Order{
			ID:         uuid.NewString(),
			UserID:     ctx.Get("userID"),
			Number:     orderNumber,
			Status:     "NEW",
			UploadedAt: time.Now(),
		}
		err = c.service.CreateOrder(ctx.Context(), order)
		if err != nil {
			switch {
			case errors.Is(err, apperrors.ErrOrderExists):
				ctx.Status(http.StatusOK)
				return ctx.JSON(ErrorResponse(err))
			case errors.Is(err, apperrors.ErrOrderOwnedByAnotherUser):
				ctx.Status(http.StatusConflict)
				return ctx.JSON(ErrorResponse(err))

			default:
				ctx.Status(http.StatusInternalServerError)
				return ctx.JSON(ErrorResponse(err))
			}
		}

		ctx.Status(http.StatusAccepted)
		return nil
	}
}

func (c *Controller) GetOrders() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		orders, err := c.service.GetOrders(ctx.Context(), ctx.Get("userID"))
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return ctx.JSON(ErrorResponse(err))
		}

		if len(orders) == 0 {
			ctx.Status(http.StatusNoContent)
			return ctx.JSON(errors.New("no orders"))
		}

		return ctx.JSON(fromService(orders))
	}
}

func fromService(orders []entity.Order) []OrderResponse {
	var res []OrderResponse
	for _, order := range orders {
		res = append(res, OrderResponse{
			Number:     order.Number,
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
			Status:     order.Status,
			Accrual:    order.Accrual,
		})
	}
	return res
}

package orders

import (
	"context"
	"errors"
	"gmart/internal/domain"
	"gmart/internal/dto"
	"gmart/internal/model/auth"
	"gmart/internal/model/luhn"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
)

// ------------ Orders upload ------------

// ordersUploadInput DTO
type ordersUploadInput struct {
	dto.AuthInput
	RawBody []byte `contentType:"text/plain" minLength:"1" doc:"Номер заказа в формате строки"`
}

type orderUploadResponse struct {
	Status int `status:"202"`
	Body   struct {
		Code   int    `json:"code" example:"200"  doc:"Код выполнения"`
		Status string `json:"status" example:"Accepted" doc:"Статус операции"`
	}
}

// signUpHandler возвращает HUMA хэндлер регистрации нового пользователя
func (o *Orders) ordersUploadHandler() func(ctx context.Context, in *ordersUploadInput) (*orderUploadResponse, error) {

	return func(ctx context.Context, in *ordersUploadInput) (*orderUploadResponse, error) {

		userID, ok := ctx.Value(auth.UserID).(domain.UserID)
		if !ok {
			return nil, huma.Error401Unauthorized("пользователь не аутентифицирован")
		}

		orderNumber := domain.OrderNumber(in.RawBody)
		if !luhn.IsValid(orderNumber) {
			return nil, huma.Error422UnprocessableEntity("номер заказа не валидируется алгоритмом luhn")
		}

		err := o.Upload(ctx, userID, orderNumber)

		if errors.Is(err, ErrOrderAlreadyUploaded) {
			resp := &orderUploadResponse{Status: 200}
			resp.Body.Code = 200
			resp.Body.Status = "AlreadyUploaded"
			return resp, nil
		}
		if errors.Is(err, ErrOrderConflict) {
			return nil, huma.Error409Conflict("номер заказа уже был загружен другим пользователем")
		}
		if err != nil {
			slog.Error("order uploaded fail",
				slog.Any("user_id", userID),
				slog.Any("order", in.RawBody),
			)
			return nil, huma.Error500InternalServerError("Внутренняя ошибка сервера")
		}

		slog.Info("order uploaded successfully",
			slog.Any("user_id", userID),
			slog.Any("order", in.RawBody),
		)

		resp := &orderUploadResponse{Status: 202}
		resp.Body.Code = 202
		resp.Body.Status = "Accepted"
		return resp, nil
	}

}

// ------------ Orders list ------------

// ordersUploadInput DTO
type ordersListInput struct {
	dto.AuthInput
}

type orderListResponse struct {
	Status int `status:"200"`
	Body   []dto.OrderItem
}

// signUpHandler возвращает HUMA хэндлер регистрации нового пользователя
func (o *Orders) ordersListHandler() func(ctx context.Context, in *ordersListInput) (*orderListResponse, error) {

	return func(ctx context.Context, in *ordersListInput) (*orderListResponse, error) {

		userID, ok := ctx.Value(auth.UserID).(domain.UserID)
		if !ok {
			return nil, huma.Error401Unauthorized("пользователь не аутентифицирован")
		}

		res, err := o.List(ctx, userID)
		if errors.Is(err, ErrOrderListIsEmpty) {
			return &orderListResponse{
				Status: 204,
			}, nil
		}
		if err != nil {
			slog.Info("order list fail",
				slog.Any("user_id", userID),
				slog.Any("err", err),
			)
			return nil, huma.Error500InternalServerError("Внутренняя ошибка сервера")
		}

		return &orderListResponse{
			Status: 200,
			Body:   res,
		}, nil
	}

}

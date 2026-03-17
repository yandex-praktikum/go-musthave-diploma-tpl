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
	Body domain.OrderNumber `minLength:"1" doc:"Номер заказа в формате строки"`
}

// Реализуем Resolver, чтобы проверять номер заказа прямо при парсинге входа
func (i *ordersUploadInput) Resolve(ctx huma.Context, prefix string) []error {
	if !luhn.IsValid(i.Body) {
		return []error{huma.Error422UnprocessableEntity("неверный формат номера заказа")}
	}
	return nil
}

type orderUploadResponse struct {
	Status int `status:"202"`
}

// signUpHandler возвращает HUMA хэндлер регистрации нового пользователя
func (o *Orders) ordersUploadHandler() func(ctx context.Context, in *ordersUploadInput) (*orderUploadResponse, error) {

	return func(ctx context.Context, in *ordersUploadInput) (*orderUploadResponse, error) {

		userID, ok := ctx.Value(auth.UserID).(domain.UserID)
		if !ok {
			return nil, huma.Error401Unauthorized("пользователь не аутентифицирован")
		}

		err := o.Upload(ctx, userID, in.Body)

		if errors.Is(err, ErrOrderAlreadyUploaded) {
			return &orderUploadResponse{Status: 200}, nil
		}
		if errors.Is(err, ErrOrderConflict) {
			return nil, huma.Error409Conflict("номер заказа уже был загружен другим пользователем")
		}
		if err != nil {
			slog.Error("order uploaded fail",
				slog.Any("user_id", userID),
				slog.Any("order", in.Body),
			)
			return nil, huma.Error500InternalServerError("Внутренняя ошибка сервера")
		}

		slog.Info("order uploaded successfully",
			slog.Any("user_id", userID),
			slog.Any("order", in.Body),
		)

		return &orderUploadResponse{Status: 202}, nil
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

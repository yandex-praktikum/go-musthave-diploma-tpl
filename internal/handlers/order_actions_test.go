package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KalessinD/gophermart/internal/common"
	"github.com/KalessinD/gophermart/internal/handlers"
	"github.com/KalessinD/gophermart/internal/middleware"
	model "github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/services/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func getTestContext(t *testing.T) context.Context {
	t.Helper()
	logger := zap.NewNop()
	return context.WithValue(t.Context(), middleware.LoggerKey, logger)
}

func TestOrdersdHandler_AddOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockOrderServiceInterface(ctrl)
	handler := handlers.NewOrdersHandler(mockService)

	tests := []struct {
		name         string
		orderNumber  string
		contentType  string
		mockBehavior func()
		expectedCode int
	}{
		{
			name:         "Success - New Order Accepted",
			orderNumber:  "12345678903", // Валидный номер
			contentType:  common.TextPlainContentType,
			expectedCode: http.StatusAccepted,
			mockBehavior: func() {
				// Ожидаем, что Store вызовется с этим номером и вернет nil
				mockService.EXPECT().Store(gomock.Any(), "12345678903").Return(nil)
			},
		},
		{
			name:         "Error - Wrong Content Type",
			orderNumber:  "123",
			contentType:  "application/json", // Неправильный Content-Type
			expectedCode: http.StatusBadRequest,
			mockBehavior: func() {
				// Store не должен вызываться
			},
		},
		{
			name:         "Error - Wrong Order Format (Luhn fail)",
			orderNumber:  "123", // Невалидный номер
			contentType:  common.TextPlainContentType,
			expectedCode: http.StatusUnprocessableEntity,
			mockBehavior: func() {
				mockService.EXPECT().Store(gomock.Any(), "123").Return(model.ErrOrderWrongFormat)
			},
		},
		{
			name:         "Error - Order Already Exists (Same User)",
			orderNumber:  "12345678903",
			contentType:  common.TextPlainContentType,
			expectedCode: http.StatusOK, // Логика хендлера: 200 для ErrOrderExists
			mockBehavior: func() {
				mockService.EXPECT().Store(gomock.Any(), "12345678903").Return(model.ErrOrderExists)
			},
		},
		{
			name:         "Error - Order Belongs To Other User",
			orderNumber:  "12345678903",
			contentType:  common.TextPlainContentType,
			expectedCode: http.StatusConflict,
			mockBehavior: func() {
				mockService.EXPECT().Store(gomock.Any(), "12345678903").Return(model.ErrOrderBelongsToOtherUser)
			},
		},
		{
			name:         "Error - Internal Server Error",
			orderNumber:  "12345678903",
			contentType:  common.TextPlainContentType,
			expectedCode: http.StatusInternalServerError,
			mockBehavior: func() {
				mockService.EXPECT().Store(gomock.Any(), "12345678903").Return(errors.New("db connection lost"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			body := strings.NewReader(tt.orderNumber)
			req, _ := http.NewRequestWithContext(getTestContext(t), http.MethodPost, "/api/user/orders", body)
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()

			handler.AddOrder(rr, req)

			require.Equal(t, tt.expectedCode, rr.Code, "Неверный HTTP статус")
		})
	}
}

func TestOrdersdHandler_ListOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockOrderServiceInterface(ctrl)
	handler := handlers.NewOrdersHandler(mockService)

	tests := []struct {
		name         string
		mockBehavior func()
		expectedCode int
		checkBody    bool
	}{
		{
			name: "Success - List Orders",
			mockBehavior: func() {
				orders := model.OrdersList{
					&model.Order{ID: "123", Status: model.OrderNewStatus},
				}
				mockService.EXPECT().List(gomock.Any()).Return(orders, nil)
			},
			expectedCode: http.StatusOK,
			checkBody:    true,
		},
		{
			name: "Success - No Orders (No Content)",
			mockBehavior: func() {
				// Сервис возвращает ErrOrderNotFound, хендлер должен вернуть 204
				mockService.EXPECT().List(gomock.Any()).Return(nil, model.ErrOrderNotFound)
			},
			expectedCode: http.StatusNoContent,
			checkBody:    false,
		},
		{
			name: "Error - Internal Server Error",
			mockBehavior: func() {
				mockService.EXPECT().List(gomock.Any()).Return(nil, errors.New("db error"))
			},
			expectedCode: http.StatusInternalServerError,
			checkBody:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			req, _ := http.NewRequestWithContext(getTestContext(t), http.MethodGet, "/api/user/orders", nil)
			rr := httptest.NewRecorder()

			handler.ListOrders(rr, req)

			require.Equal(t, tt.expectedCode, rr.Code, "Неверный HTTP статус")

			if tt.checkBody {
				// Проверяем, что вернулся JSON массив
				require.Contains(t, rr.Header().Get("Content-Type"), "application/json")
				require.JSONEq(t, `[{"number":"123","status":"NEW","uploaded":"0001-01-01T00:00:00Z"}]`, rr.Body.String())
			}
		})
	}
}

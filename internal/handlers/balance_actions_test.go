package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KalessinD/gophermart/internal/common"
	"github.com/KalessinD/gophermart/internal/handlers"
	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestBalanceHandler_GetLoyalityBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockBalanceServiceInterface(ctrl)
	handler := handlers.NewBalancesHandler(mockService)

	t.Run("success", func(t *testing.T) {
		ctx := getTestContext(t)
		balance := models.NewBalanceItem(500, 100) // 5.00 и 1.00
		mockService.EXPECT().GetBalanceInfo(ctx).Return(balance, nil)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/user/balance", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.GetLoyalityBalance(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, common.AppJSONContentType, rr.Header().Get("Content-Type"))

		expectedBody, _ := balance.ToJSON()
		assert.JSONEq(t, string(expectedBody), rr.Body.String())
	})

	t.Run("internal error", func(t *testing.T) {
		ctx := getTestContext(t)
		mockService.EXPECT().GetBalanceInfo(ctx).Return(nil, errors.New("db error"))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/user/balance", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.GetLoyalityBalance(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestBalanceHandler_WithdrawBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockBalanceServiceInterface(ctrl)
	handler := handlers.NewBalancesHandler(mockService)

	validOrderID := "12345678903" // Валидный по Луну номер для контекста (если сервис проверяет)
	input := models.Withdrawn{OrderID: validOrderID, Sum: 100}
	inputBody, _ := json.Marshal(input)

	t.Run("success", func(t *testing.T) {
		ctx := getTestContext(t)
		// Ожидаем, что сервис получит структуру с суммой
		mockService.EXPECT().Withdraw(ctx, gomock.Cond(func(x any) bool {
			w, ok := x.(*models.Withdrawn)
			return ok && w.OrderID == validOrderID && w.Sum == 100
		})).Return(nil)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(inputBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", common.AppJSONContentType)

		rr := httptest.NewRecorder()
		handler.WithdrawBalance(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("wrong content type", func(t *testing.T) {
		ctx := getTestContext(t)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(inputBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "text/plain")

		rr := httptest.NewRecorder()
		handler.WithdrawBalance(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid json body", func(t *testing.T) {
		ctx := getTestContext(t)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader([]byte("{bad json}")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", common.AppJSONContentType)

		rr := httptest.NewRecorder()
		handler.WithdrawBalance(rr, req)

		// Код возвращает 500 при ошибке парсинга JSON
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("wrong order format (422)", func(t *testing.T) {
		ctx := getTestContext(t)
		mockService.EXPECT().Withdraw(ctx, gomock.Any()).Return(models.ErrOrderWrongFormat)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(inputBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", common.AppJSONContentType)

		rr := httptest.NewRecorder()
		handler.WithdrawBalance(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("not enough funds (402)", func(t *testing.T) {
		ctx := getTestContext(t)
		mockService.EXPECT().Withdraw(ctx, gomock.Any()).Return(models.ErrUserBalanceIsNotEnough)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(inputBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", common.AppJSONContentType)

		rr := httptest.NewRecorder()
		handler.WithdrawBalance(rr, req)

		assert.Equal(t, http.StatusPaymentRequired, rr.Code)
	})
}

func TestBalanceHandler_ListWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockBalanceServiceInterface(ctrl)
	handler := handlers.NewBalancesHandler(mockService)

	t.Run("success with data", func(t *testing.T) {
		ctx := getTestContext(t)
		list := models.WithdrawnList{
			{OrderID: "1", Sum: 100},
		}
		mockService.EXPECT().ListWithdrawals(ctx).Return(list, nil)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/user/withdrawals", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ListWithdrawals(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, common.AppJSONContentType, rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Body.String(), `"order":"1"`)
	})

	t.Run("success empty list", func(t *testing.T) {
		ctx := getTestContext(t)

		mockService.EXPECT().ListWithdrawals(ctx).Return(models.WithdrawnList{}, nil)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/user/withdrawals", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ListWithdrawals(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.Equal(t, "", rr.Body.String())
	})

	t.Run("internal error", func(t *testing.T) {
		ctx := getTestContext(t)
		mockService.EXPECT().ListWithdrawals(ctx).Return(nil, errors.New("db error"))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/user/withdrawals", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ListWithdrawals(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/auth"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/config"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/handler"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/models"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository/mock"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/router"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
)

func setupRouterWithBalance(t *testing.T, ctrl *gomock.Controller) (*gin.Engine, *mock.MockOrderRepository, *mock.MockWithdrawalRepository) {
	t.Helper()
	orderRepo := mock.NewMockOrderRepository(ctrl)
	withdrawalRepo := mock.NewMockWithdrawalRepository(ctrl)
	services := &service.Services{
		User:    nil,
		Order:   nil,
		Balance: service.NewBalanceService(orderRepo, withdrawalRepo),
	}
	h := handler.New(services, testCookieSecret)
	cfg := &config.Config{CookieSecret: testCookieSecret}
	return router.SetupRouter(h, cfg), orderRepo, withdrawalRepo
}

func TestHandler_GetBalance_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, orderRepo, withdrawalRepo := setupRouterWithBalance(t, ctrl)
	userID := int64(1)

	orderRepo.EXPECT().
		GetTotalAccrualsByUserID(gomock.Any(), userID).
		Return(int64(500), nil)
	withdrawalRepo.EXPECT().
		GetTotalWithdrawnByUserID(gomock.Any(), userID).
		Return(int64(100), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req.AddCookie(auth.NewCookie(userID, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetBalance: status = %d; want %d", w.Code, http.StatusOK)
		return
	}
	var body handler.BalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("GetBalance decode: %v", err)
	}
	if body.Current != 400 || body.Withdrawn != 100 {
		t.Errorf("GetBalance: current=%f withdrawn=%f; want 400, 100", body.Current, body.Withdrawn)
	}
}

func TestHandler_GetBalance_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, _ := setupRouterWithBalance(t, ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetBalance no cookie: status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_GetBalance_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, orderRepo, _ := setupRouterWithBalance(t, ctrl)

	orderRepo.EXPECT().
		GetTotalAccrualsByUserID(gomock.Any(), int64(1)).
		Return(int64(0), errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req.AddCookie(auth.NewCookie(1, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("GetBalance internal: status = %d; want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandler_PostWithdraw_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, withdrawalRepo := setupRouterWithBalance(t, ctrl)
	userID := int64(1)

	withdrawalRepo.EXPECT().
		Withdraw(gomock.Any(), userID, validLuhnNumber, int64(100)).
		Return(nil)

	body := handler.WithdrawRequest{Order: validLuhnNumber, Sum: 100}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(auth.NewCookie(userID, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("PostWithdraw: status = %d; want %d", w.Code, http.StatusOK)
	}
}

func TestHandler_PostWithdraw_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, _ := setupRouterWithBalance(t, ctrl)

	body := handler.WithdrawRequest{Order: validLuhnNumber, Sum: 100}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("PostWithdraw no cookie: status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_PostWithdraw_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, _ := setupRouterWithBalance(t, ctrl)

	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(auth.NewCookie(1, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("PostWithdraw invalid JSON: status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_PostWithdraw_SumNotPositive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, _ := setupRouterWithBalance(t, ctrl)

	body := handler.WithdrawRequest{Order: validLuhnNumber, Sum: 0}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(auth.NewCookie(1, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("PostWithdraw sum 0: status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_PostWithdraw_SumFractional(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, _ := setupRouterWithBalance(t, ctrl)

	body := handler.WithdrawRequest{Order: validLuhnNumber, Sum: 50.5}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(auth.NewCookie(1, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("PostWithdraw sum fractional: status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_PostWithdraw_InsufficientFunds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, withdrawalRepo := setupRouterWithBalance(t, ctrl)
	userID := int64(1)

	withdrawalRepo.EXPECT().
		Withdraw(gomock.Any(), userID, validLuhnNumber, int64(1000)).
		Return(&repository.ErrInsufficientFunds{Order: validLuhnNumber})

	body := handler.WithdrawRequest{Order: validLuhnNumber, Sum: 1000}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(auth.NewCookie(userID, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusPaymentRequired {
		t.Errorf("PostWithdraw insufficient: status = %d; want %d (402)", w.Code, http.StatusPaymentRequired)
	}
}

func TestHandler_PostWithdraw_InvalidOrderLuhn(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, _ := setupRouterWithBalance(t, ctrl)

	body := handler.WithdrawRequest{Order: "1234567890", Sum: 100}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(auth.NewCookie(1, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("PostWithdraw invalid Luhn: status = %d; want %d", w.Code, http.StatusUnprocessableEntity)
	}
}

func TestHandler_PostWithdraw_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, withdrawalRepo := setupRouterWithBalance(t, ctrl)

	withdrawalRepo.EXPECT().
		Withdraw(gomock.Any(), int64(1), validLuhnNumber, int64(100)).
		Return(errors.New("db error"))

	body := handler.WithdrawRequest{Order: validLuhnNumber, Sum: 100}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(auth.NewCookie(1, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("PostWithdraw internal: status = %d; want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandler_GetWithdrawals_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, withdrawalRepo := setupRouterWithBalance(t, ctrl)
	userID := int64(1)
	ts := time.Now()
	list := []*models.Withdrawal{
		{ID: 1, UserID: userID, Order: "79927398713", Sum: 100, ProcessedAt: ts},
		{ID: 2, UserID: userID, Order: "2377225624", Sum: 50, ProcessedAt: ts.Add(-time.Hour)},
	}
	withdrawalRepo.EXPECT().
		ListByUserID(gomock.Any(), userID).
		Return(list, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req.AddCookie(auth.NewCookie(userID, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetWithdrawals: status = %d; want %d", w.Code, http.StatusOK)
		return
	}
	var items []handler.WithdrawalItem
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("GetWithdrawals decode: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("GetWithdrawals: len = %d; want 2", len(items))
	}
	if items[0].Order != "79927398713" || items[0].Sum != 100 {
		t.Errorf("GetWithdrawals first: order=%q sum=%f", items[0].Order, items[0].Sum)
	}
}

func TestHandler_GetWithdrawals_NoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, withdrawalRepo := setupRouterWithBalance(t, ctrl)
	userID := int64(1)

	withdrawalRepo.EXPECT().
		ListByUserID(gomock.Any(), userID).
		Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req.AddCookie(auth.NewCookie(userID, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("GetWithdrawals empty: status = %d; want %d", w.Code, http.StatusNoContent)
	}
}

func TestHandler_GetWithdrawals_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, _ := setupRouterWithBalance(t, ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetWithdrawals no cookie: status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_GetWithdrawals_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, withdrawalRepo := setupRouterWithBalance(t, ctrl)

	withdrawalRepo.EXPECT().
		ListByUserID(gomock.Any(), int64(1)).
		Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req.AddCookie(auth.NewCookie(1, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("GetWithdrawals internal: status = %d; want %d", w.Code, http.StatusInternalServerError)
	}
}

package handler_test

import (
	"bytes"
	"encoding/json"
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

const testCookieSecret = "test-secret"
const validLuhnNumber = "79927398713"

func setupRouterWithOrders(t *testing.T, ctrl *gomock.Controller) (*gin.Engine, *mock.MockUserRepository, *mock.MockOrderRepository) {
	t.Helper()
	userRepo := mock.NewMockUserRepository(ctrl)
	orderRepo := mock.NewMockOrderRepository(ctrl)
	services := &service.Services{
		User:  service.NewUserService(userRepo),
		Order: service.NewOrderService(orderRepo),
	}
	h := handler.New(services, testCookieSecret)
	cfg := &config.Config{CookieSecret: testCookieSecret}
	return router.SetupRouter(h, cfg), userRepo, orderRepo
}

func TestHandler_PostOrders_SuccessNew(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, orderRepo := setupRouterWithOrders(t, ctrl)
	userID := int64(1)

	orderRepo.EXPECT().
		GetByNumber(gomock.Any(), validLuhnNumber).
		Return(nil, &repository.ErrOrderNotFound{Number: validLuhnNumber})
	orderRepo.EXPECT().
		Create(gomock.Any(), userID, validLuhnNumber, service.OrderStatusNew).
		Return(&models.Order{ID: 1, UserID: userID, Number: validLuhnNumber, Status: service.OrderStatusNew, UploadedAt: time.Now()}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(validLuhnNumber)))
	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(auth.NewCookie(userID, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("PostOrders new: status = %d; want %d", w.Code, http.StatusAccepted)
	}
}

func TestHandler_PostOrders_SuccessAlreadyUploaded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, orderRepo := setupRouterWithOrders(t, ctrl)
	userID := int64(1)

	existing := &models.Order{ID: 1, UserID: userID, Number: validLuhnNumber, Status: service.OrderStatusProcessing, UploadedAt: time.Now()}
	orderRepo.EXPECT().
		GetByNumber(gomock.Any(), validLuhnNumber).
		Return(existing, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(validLuhnNumber)))
	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(auth.NewCookie(userID, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("PostOrders already uploaded: status = %d; want %d", w.Code, http.StatusOK)
	}
}

func TestHandler_PostOrders_UnauthorizedNoCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, _ := setupRouterWithOrders(t, ctrl)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(validLuhnNumber)))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("PostOrders no cookie: status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_PostOrders_ConflictOtherUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, orderRepo := setupRouterWithOrders(t, ctrl)
	reqUserID := int64(2)

	existing := &models.Order{ID: 1, UserID: 1, Number: validLuhnNumber, Status: service.OrderStatusNew, UploadedAt: time.Now()}
	orderRepo.EXPECT().
		GetByNumber(gomock.Any(), validLuhnNumber).
		Return(existing, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(validLuhnNumber)))
	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(auth.NewCookie(reqUserID, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("PostOrders other user: status = %d; want %d", w.Code, http.StatusConflict)
	}
}

func TestHandler_PostOrders_UnprocessableInvalidNumber(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, _ := setupRouterWithOrders(t, ctrl)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("1234abc")))
	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(auth.NewCookie(1, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("PostOrders invalid number: status = %d; want %d", w.Code, http.StatusUnprocessableEntity)
	}
}

func TestHandler_PostOrders_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, orderRepo := setupRouterWithOrders(t, ctrl)

	orderRepo.EXPECT().
		GetByNumber(gomock.Any(), validLuhnNumber).
		Return(nil, http.ErrHandlerTimeout)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(validLuhnNumber)))
	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(auth.NewCookie(1, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("PostOrders internal: status = %d; want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandler_GetOrders_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, orderRepo := setupRouterWithOrders(t, ctrl)
	userID := int64(1)
	accrual := 100
	list := []*models.Order{
		{ID: 1, UserID: userID, Number: "111", Status: service.OrderStatusProcessed, Accrual: &accrual, UploadedAt: time.Now()},
		{ID: 2, UserID: userID, Number: "222", Status: service.OrderStatusNew, UploadedAt: time.Now()},
	}
	orderRepo.EXPECT().
		ListByUserID(gomock.Any(), userID).
		Return(list, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.AddCookie(auth.NewCookie(userID, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetOrders: status = %d; want %d", w.Code, http.StatusOK)
		return
	}
	var items []handler.OrderItem
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("GetOrders decode: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("GetOrders: len = %d; want 2", len(items))
	}
	if items[0].Number != "111" || items[1].Number != "222" {
		t.Errorf("GetOrders: numbers = %q, %q", items[0].Number, items[1].Number)
	}
}

func TestHandler_GetOrders_NoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, orderRepo := setupRouterWithOrders(t, ctrl)
	userID := int64(1)

	orderRepo.EXPECT().
		ListByUserID(gomock.Any(), userID).
		Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.AddCookie(auth.NewCookie(userID, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("GetOrders empty: status = %d; want %d", w.Code, http.StatusNoContent)
	}
}

func TestHandler_GetOrders_UnauthorizedNoCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, _ := setupRouterWithOrders(t, ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetOrders no cookie: status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_GetOrders_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _, orderRepo := setupRouterWithOrders(t, ctrl)

	orderRepo.EXPECT().
		ListByUserID(gomock.Any(), int64(1)).
		Return(nil, http.ErrHandlerTimeout)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.AddCookie(auth.NewCookie(1, testCookieSecret))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("GetOrders internal: status = %d; want %d", w.Code, http.StatusInternalServerError)
	}
}

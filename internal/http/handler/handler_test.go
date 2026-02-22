package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"

	"github.com/anon-d/gophermarket/internal/domain"
	"github.com/anon-d/gophermarket/internal/http/middleware"
	"github.com/anon-d/gophermarket/internal/http/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// errorReadCloser имитирует ошибку чтения тела запроса
type errorReadCloser struct{}

func (e errorReadCloser) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func (e errorReadCloser) Close() error {
	return nil
}

// --- helpers ---

func newTestContext(method, path string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, body)
	return c, w
}

func newAuthContext(method, path string, body io.Reader, userID string) (*gin.Context, *httptest.ResponseRecorder) {
	c, w := newTestContext(method, path, body)
	c.Set(middleware.UserIDKey, userID)
	return c, w
}

func assertStatus(t *testing.T, c *gin.Context, want int) {
	t.Helper()
	c.Writer.WriteHeaderNow()
	got := c.Writer.Status()
	if got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
}

// --- TestNewGopherHandler ---

func TestNewGopherHandler(t *testing.T) {
	h := NewGopherHandler(nil)
	if h == nil {
		t.Fatal("NewGopherHandler() returned nil")
	}
}

// --- RegisterUser ---

func TestRegisterUser_BadJSON(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newTestContext(http.MethodPost, "/api/user/register", strings.NewReader("{"))

	h.RegisterUser(c)
	assertStatus(t, c, http.StatusBadRequest)
}

func TestRegisterUser_MissingFields(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newTestContext(http.MethodPost, "/api/user/register", strings.NewReader(`{"login":"u"}`))

	h.RegisterUser(c)
	assertStatus(t, c, http.StatusBadRequest)
}

func TestRegisterUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		RegisterUser(gomock.Any(), "user1", "pass1").
		Return("jwt-token-123", nil)

	h := NewGopherHandler(mockSvc)
	c, w := newTestContext(http.MethodPost, "/api/user/register",
		strings.NewReader(`{"login":"user1","password":"pass1"}`))

	h.RegisterUser(c)
	assertStatus(t, c, http.StatusOK)

	if auth := w.Header().Get("Authorization"); auth != "Bearer jwt-token-123" {
		t.Errorf("Authorization = %q, want %q", auth, "Bearer jwt-token-123")
	}
}

func TestRegisterUser_Conflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		RegisterUser(gomock.Any(), "existing", "pass").
		Return("", service.ErrUserExists)

	h := NewGopherHandler(mockSvc)
	c, _ := newTestContext(http.MethodPost, "/api/user/register",
		strings.NewReader(`{"login":"existing","password":"pass"}`))

	h.RegisterUser(c)
	assertStatus(t, c, http.StatusConflict)
}

func TestRegisterUser_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		RegisterUser(gomock.Any(), "user1", "pass1").
		Return("", errors.New("db error"))

	h := NewGopherHandler(mockSvc)
	c, _ := newTestContext(http.MethodPost, "/api/user/register",
		strings.NewReader(`{"login":"user1","password":"pass1"}`))

	h.RegisterUser(c)
	assertStatus(t, c, http.StatusInternalServerError)
}

// --- LoginUser ---

func TestLoginUser_BadJSON(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newTestContext(http.MethodPost, "/api/user/login", strings.NewReader("not-json"))

	h.LoginUser(c)
	assertStatus(t, c, http.StatusBadRequest)
}

func TestLoginUser_MissingFields(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newTestContext(http.MethodPost, "/api/user/login", strings.NewReader(`{"password":"p"}`))

	h.LoginUser(c)
	assertStatus(t, c, http.StatusBadRequest)
}

func TestLoginUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		LoginUser(gomock.Any(), "user1", "pass1").
		Return("jwt-token-456", nil)

	h := NewGopherHandler(mockSvc)
	c, w := newTestContext(http.MethodPost, "/api/user/login",
		strings.NewReader(`{"login":"user1","password":"pass1"}`))

	h.LoginUser(c)
	assertStatus(t, c, http.StatusOK)

	if auth := w.Header().Get("Authorization"); auth != "Bearer jwt-token-456" {
		t.Errorf("Authorization = %q, want %q", auth, "Bearer jwt-token-456")
	}
}

func TestLoginUser_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		LoginUser(gomock.Any(), "user1", "wrong").
		Return("", service.ErrInvalidCredentials)

	h := NewGopherHandler(mockSvc)
	c, _ := newTestContext(http.MethodPost, "/api/user/login",
		strings.NewReader(`{"login":"user1","password":"wrong"}`))

	h.LoginUser(c)
	assertStatus(t, c, http.StatusUnauthorized)
}

func TestLoginUser_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		LoginUser(gomock.Any(), "user1", "pass1").
		Return("", errors.New("db error"))

	h := NewGopherHandler(mockSvc)
	c, _ := newTestContext(http.MethodPost, "/api/user/login",
		strings.NewReader(`{"login":"user1","password":"pass1"}`))

	h.LoginUser(c)
	assertStatus(t, c, http.StatusInternalServerError)
}

// --- CreateOrder ---

func TestCreateOrder_Unauthorized(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newTestContext(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"))

	h.CreateOrder(c)
	assertStatus(t, c, http.StatusUnauthorized)
}

func TestCreateOrder_BodyReadError(t *testing.T) {
	h := NewGopherHandler(nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", nil)
	req.Body = errorReadCloser{}
	c.Request = req
	c.Set(middleware.UserIDKey, "user-1")

	h.CreateOrder(c)
	assertStatus(t, c, http.StatusBadRequest)
}

func TestCreateOrder_EmptyOrderNumber(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newAuthContext(http.MethodPost, "/api/user/orders", strings.NewReader("   \n\t"), "user-1")

	h.CreateOrder(c)
	assertStatus(t, c, http.StatusBadRequest)
}

func TestCreateOrder_Accepted(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		CreateOrder(gomock.Any(), "user-1", "12345678903").
		Return(nil)

	h := NewGopherHandler(mockSvc)
	c, _ := newAuthContext(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"), "user-1")

	h.CreateOrder(c)
	assertStatus(t, c, http.StatusAccepted)
}

func TestCreateOrder_AlreadyExistsSameUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		CreateOrder(gomock.Any(), "user-1", "12345678903").
		Return(service.ErrOrderExists)

	h := NewGopherHandler(mockSvc)
	c, _ := newAuthContext(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"), "user-1")

	h.CreateOrder(c)
	assertStatus(t, c, http.StatusOK)
}

func TestCreateOrder_AlreadyExistsAnotherUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		CreateOrder(gomock.Any(), "user-1", "12345678903").
		Return(service.ErrOrderExistsByAnotherUser)

	h := NewGopherHandler(mockSvc)
	c, _ := newAuthContext(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"), "user-1")

	h.CreateOrder(c)
	assertStatus(t, c, http.StatusConflict)
}

func TestCreateOrder_InvalidNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		CreateOrder(gomock.Any(), "user-1", "bad-number").
		Return(service.ErrInvalidOrderNumber)

	h := NewGopherHandler(mockSvc)
	c, _ := newAuthContext(http.MethodPost, "/api/user/orders", strings.NewReader("bad-number"), "user-1")

	h.CreateOrder(c)
	assertStatus(t, c, http.StatusUnprocessableEntity)
}

func TestCreateOrder_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		CreateOrder(gomock.Any(), "user-1", "12345678903").
		Return(errors.New("db error"))

	h := NewGopherHandler(mockSvc)
	c, _ := newAuthContext(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"), "user-1")

	h.CreateOrder(c)
	assertStatus(t, c, http.StatusInternalServerError)
}

// --- GetOrders ---

func TestGetOrders_Unauthorized(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newTestContext(http.MethodGet, "/api/user/orders", nil)

	h.GetOrders(c)
	assertStatus(t, c, http.StatusUnauthorized)
}

func TestGetOrders_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	orders := []domain.Order{
		{Number: "111", Status: domain.OrderStatusProcessed, Accrual: 500, UploadedAt: now},
		{Number: "222", Status: domain.OrderStatusNew, Accrual: 0, UploadedAt: now},
	}

	mockSvc.EXPECT().
		GetOrders(gomock.Any(), "user-1").
		Return(orders, nil)

	h := NewGopherHandler(mockSvc)
	c, w := newAuthContext(http.MethodGet, "/api/user/orders", nil, "user-1")

	h.GetOrders(c)
	assertStatus(t, c, http.StatusOK)

	var resp []OrderResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("len(response) = %d, want 2", len(resp))
	}
	if resp[0].Number != "111" {
		t.Errorf("order[0].Number = %q, want %q", resp[0].Number, "111")
	}
	if resp[0].Accrual == nil || *resp[0].Accrual != 500 {
		t.Errorf("order[0].Accrual = %v, want 500", resp[0].Accrual)
	}
	if resp[1].Accrual != nil {
		t.Errorf("order[1].Accrual = %v, want nil", resp[1].Accrual)
	}
}

func TestGetOrders_NoContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		GetOrders(gomock.Any(), "user-1").
		Return([]domain.Order{}, nil)

	h := NewGopherHandler(mockSvc)
	c, _ := newAuthContext(http.MethodGet, "/api/user/orders", nil, "user-1")

	h.GetOrders(c)
	assertStatus(t, c, http.StatusNoContent)
}

func TestGetOrders_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		GetOrders(gomock.Any(), "user-1").
		Return(nil, errors.New("db error"))

	h := NewGopherHandler(mockSvc)
	c, _ := newAuthContext(http.MethodGet, "/api/user/orders", nil, "user-1")

	h.GetOrders(c)
	assertStatus(t, c, http.StatusInternalServerError)
}

// --- GetBalance ---

func TestGetBalance_Unauthorized(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newTestContext(http.MethodGet, "/api/user/balance", nil)

	h.GetBalance(c)
	assertStatus(t, c, http.StatusUnauthorized)
}

func TestGetBalance_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		GetBalance(gomock.Any(), "user-1").
		Return(&domain.Balance{Current: 100.5, Withdrawn: 50.0}, nil)

	h := NewGopherHandler(mockSvc)
	c, w := newAuthContext(http.MethodGet, "/api/user/balance", nil, "user-1")

	h.GetBalance(c)
	assertStatus(t, c, http.StatusOK)

	var resp BalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Current != 100.5 {
		t.Errorf("Current = %v, want 100.5", resp.Current)
	}
	if resp.Withdrawn != 50.0 {
		t.Errorf("Withdrawn = %v, want 50.0", resp.Withdrawn)
	}
}

func TestGetBalance_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		GetBalance(gomock.Any(), "user-1").
		Return(nil, errors.New("db error"))

	h := NewGopherHandler(mockSvc)
	c, _ := newAuthContext(http.MethodGet, "/api/user/balance", nil, "user-1")

	h.GetBalance(c)
	assertStatus(t, c, http.StatusInternalServerError)
}

// --- WithdrawBalance ---

func TestWithdrawBalance_Unauthorized(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newTestContext(http.MethodPost, "/api/user/balance/withdraw",
		strings.NewReader(`{"order":"123","sum":10}`))

	h.WithdrawBalance(c)
	assertStatus(t, c, http.StatusUnauthorized)
}

func TestWithdrawBalance_BadJSON(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newAuthContext(http.MethodPost, "/api/user/balance/withdraw",
		strings.NewReader("{"), "user-1")

	h.WithdrawBalance(c)
	assertStatus(t, c, http.StatusBadRequest)
}

func TestWithdrawBalance_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		Withdraw(gomock.Any(), "user-1", "12345678903", 100.0).
		Return(nil)

	h := NewGopherHandler(mockSvc)
	c, _ := newAuthContext(http.MethodPost, "/api/user/balance/withdraw",
		strings.NewReader(`{"order":"12345678903","sum":100}`), "user-1")

	h.WithdrawBalance(c)
	assertStatus(t, c, http.StatusOK)
}

func TestWithdrawBalance_InsufficientFunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		Withdraw(gomock.Any(), "user-1", "12345678903", 9999.0).
		Return(service.ErrInsufficientFunds)

	h := NewGopherHandler(mockSvc)
	c, _ := newAuthContext(http.MethodPost, "/api/user/balance/withdraw",
		strings.NewReader(`{"order":"12345678903","sum":9999}`), "user-1")

	h.WithdrawBalance(c)
	assertStatus(t, c, http.StatusPaymentRequired)
}

func TestWithdrawBalance_InvalidOrderNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		Withdraw(gomock.Any(), "user-1", "bad", 10.0).
		Return(service.ErrInvalidOrderNumber)

	h := NewGopherHandler(mockSvc)
	c, _ := newAuthContext(http.MethodPost, "/api/user/balance/withdraw",
		strings.NewReader(`{"order":"bad","sum":10}`), "user-1")

	h.WithdrawBalance(c)
	assertStatus(t, c, http.StatusUnprocessableEntity)
}

func TestWithdrawBalance_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		Withdraw(gomock.Any(), "user-1", "12345678903", 10.0).
		Return(errors.New("db error"))

	h := NewGopherHandler(mockSvc)
	c, _ := newAuthContext(http.MethodPost, "/api/user/balance/withdraw",
		strings.NewReader(`{"order":"12345678903","sum":10}`), "user-1")

	h.WithdrawBalance(c)
	assertStatus(t, c, http.StatusInternalServerError)
}

// --- GetWithdrawals ---

func TestGetWithdrawals_Unauthorized(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newTestContext(http.MethodGet, "/api/user/withdrawals", nil)

	h.GetWithdrawals(c)
	assertStatus(t, c, http.StatusUnauthorized)
}

func TestGetWithdrawals_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	withdrawals := []domain.Withdrawal{
		{OrderNumber: "111", Sum: 50.0, ProcessedAt: now},
		{OrderNumber: "222", Sum: 25.5, ProcessedAt: now},
	}

	mockSvc.EXPECT().
		GetWithdrawals(gomock.Any(), "user-1").
		Return(withdrawals, nil)

	h := NewGopherHandler(mockSvc)
	c, w := newAuthContext(http.MethodGet, "/api/user/withdrawals", nil, "user-1")

	h.GetWithdrawals(c)
	assertStatus(t, c, http.StatusOK)

	var resp []WithdrawalResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("len(response) = %d, want 2", len(resp))
	}
	if resp[0].Order != "111" || resp[0].Sum != 50.0 {
		t.Errorf("withdrawal[0] = %+v, want order=111 sum=50", resp[0])
	}
	if resp[1].Order != "222" || resp[1].Sum != 25.5 {
		t.Errorf("withdrawal[1] = %+v, want order=222 sum=25.5", resp[1])
	}
}

func TestGetWithdrawals_NoContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		GetWithdrawals(gomock.Any(), "user-1").
		Return([]domain.Withdrawal{}, nil)

	h := NewGopherHandler(mockSvc)
	c, _ := newAuthContext(http.MethodGet, "/api/user/withdrawals", nil, "user-1")

	h.GetWithdrawals(c)
	assertStatus(t, c, http.StatusNoContent)
}

func TestGetWithdrawals_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		GetWithdrawals(gomock.Any(), "user-1").
		Return(nil, errors.New("db error"))

	h := NewGopherHandler(mockSvc)
	c, _ := newAuthContext(http.MethodGet, "/api/user/withdrawals", nil, "user-1")

	h.GetWithdrawals(c)
	assertStatus(t, c, http.StatusInternalServerError)
}

// --- setAuthCookie ---

func TestSetAuthCookie_SetsHeaderAndCookie(t *testing.T) {
	h := NewGopherHandler(nil)
	c, w := newTestContext(http.MethodPost, "/api/user/login", nil)

	h.setAuthCookie(c, "test-token")

	auth := w.Header().Get("Authorization")
	if auth != "Bearer test-token" {
		t.Errorf("Authorization header = %q, want %q", auth, "Bearer test-token")
	}

	setCookie := w.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "token=test-token") {
		t.Errorf("Set-Cookie header = %q, want contains token=test-token", setCookie)
	}
}

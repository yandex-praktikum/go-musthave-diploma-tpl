package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/anon-d/gophermarket/internal/http/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type errorReadCloser struct{}

func (e errorReadCloser) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func (e errorReadCloser) Close() error {
	return nil
}

func newTestContext(method, path string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, body)
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

func TestNewGopherHandler(t *testing.T) {
	h := NewGopherHandler(nil)
	if h == nil {
		t.Fatal("NewGopherHandler() returned nil")
	}
}

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
	c, _ := newTestContext(http.MethodPost, "/api/user/orders", strings.NewReader("   \n\t"))
	c.Set(middleware.UserIDKey, "user-1")

	h.CreateOrder(c)
	assertStatus(t, c, http.StatusBadRequest)
}

func TestGetOrders_Unauthorized(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newTestContext(http.MethodGet, "/api/user/orders", nil)

	h.GetOrders(c)
	assertStatus(t, c, http.StatusUnauthorized)
}

func TestGetBalance_Unauthorized(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newTestContext(http.MethodGet, "/api/user/balance", nil)

	h.GetBalance(c)
	assertStatus(t, c, http.StatusUnauthorized)
}

func TestWithdrawBalance_Unauthorized(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newTestContext(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader(`{"order":"123","sum":10}`))

	h.WithdrawBalance(c)
	assertStatus(t, c, http.StatusUnauthorized)
}

func TestWithdrawBalance_BadJSON(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newTestContext(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader("{"))
	c.Set(middleware.UserIDKey, "user-1")

	h.WithdrawBalance(c)
	assertStatus(t, c, http.StatusBadRequest)
}

func TestGetWithdrawals_Unauthorized(t *testing.T) {
	h := NewGopherHandler(nil)
	c, _ := newTestContext(http.MethodGet, "/api/user/withdrawals", nil)

	h.GetWithdrawals(c)
	assertStatus(t, c, http.StatusUnauthorized)
}

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

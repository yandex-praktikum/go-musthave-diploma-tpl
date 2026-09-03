package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/true18/go-musthave-diploma-tpl/internal/auth"
	"github.com/true18/go-musthave-diploma-tpl/internal/handler"
	"github.com/true18/go-musthave-diploma-tpl/internal/money"
	"github.com/true18/go-musthave-diploma-tpl/internal/repository"
	"github.com/true18/go-musthave-diploma-tpl/internal/server"
)

func TestNewRoutes(t *testing.T) {
	authManager := auth.New([]byte("test-secret"))
	appHandler := handler.New(fakeService{}, authManager)
	httpServer := server.New(":0", appHandler, authManager, zap.NewNop())

	response := httptest.NewRecorder()
	httpServer.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/user/register", nil))
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("register GET status = %d", response.Code)
	}

	response = httptest.NewRecorder()
	httpServer.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/user/balance", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("private route status = %d", response.Code)
	}
}

type fakeService struct{}

func (fakeService) Register(context.Context, string, string) (int64, error) {
	return 1, nil
}

func (fakeService) Login(context.Context, string, string) (int64, error) {
	return 1, nil
}

func (fakeService) UploadOrder(context.Context, int64, string) error {
	return nil
}

func (fakeService) ListOrders(context.Context, int64) ([]repository.Order, error) {
	return nil, nil
}

func (fakeService) Balance(context.Context, int64) (repository.Balance, error) {
	return repository.Balance{}, nil
}

func (fakeService) Withdraw(context.Context, int64, string, money.Amount) error {
	return nil
}

func (fakeService) ListWithdrawals(context.Context, int64) ([]repository.Withdrawal, error) {
	return nil, nil
}

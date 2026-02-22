package worker

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/anon-d/gophermarket/internal/repository"
)

var testLogger = zap.NewNop()

// --- NewAccrualWorker ---

func TestNewAccrualWorker(t *testing.T) {
	w := NewAccrualWorker(nil, "http://localhost:8080", testLogger)
	if w == nil {
		t.Fatal("NewAccrualWorker() returned nil")
	}
	if w.interval != 5*time.Second {
		t.Errorf("interval = %v, want 5s", w.interval)
	}
	if w.accrualAddress != "http://localhost:8080" {
		t.Errorf("accrualAddress = %q, want %q", w.accrualAddress, "http://localhost:8080")
	}
}

// --- Start ---

func TestStart_ContextCancel(t *testing.T) {
	w := NewAccrualWorker(nil, "", testLogger)
	w.interval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		w.Start(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// OK — воркер остановился
	case <-time.After(time.Second):
		t.Fatal("Start did not stop after context cancel")
	}
}

// --- processOrders ---

func TestProcessOrders_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockWorkerRepository(ctrl)

	uid := uuid.New()
	orders := []repository.Order{
		{Number: "111", UserID: uid, Status: "NEW"},
	}

	// Возвращаем заказы
	mockRepo.EXPECT().
		GetOrdersForProcessing(gomock.Any()).
		Return(orders, nil)

	// Ожидаем обновление статуса (processOrder вызовет HTTP → mock-сервер → UpdateOrderStatus)
	mockRepo.EXPECT().
		UpdateOrderStatus(gomock.Any(), "111", "PROCESSED", 500.0).
		Return(nil)

	// Mock HTTP-сервер для accrual API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := AccrualResponse{Order: "111", Status: "PROCESSED", Accrual: 500.0}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	w := NewAccrualWorker(mockRepo, server.URL, testLogger)
	w.processOrders(context.Background())
}

func TestProcessOrders_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockWorkerRepository(ctrl)

	mockRepo.EXPECT().
		GetOrdersForProcessing(gomock.Any()).
		Return(nil, errors.New("db error"))

	w := NewAccrualWorker(mockRepo, "", testLogger)
	// Не паникует, просто логирует ошибку
	w.processOrders(context.Background())
}

func TestProcessOrders_ContextCancelDuringIteration(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockWorkerRepository(ctrl)

	uid := uuid.New()
	orders := []repository.Order{
		{Number: "111", UserID: uid},
		{Number: "222", UserID: uid},
	}

	mockRepo.EXPECT().
		GetOrdersForProcessing(gomock.Any()).
		Return(orders, nil)

	// Не ожидаем вызовов UpdateOrderStatus — контекст отменён

	w := NewAccrualWorker(mockRepo, "http://invalid", testLogger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // отменяем сразу

	w.processOrders(ctx)
}

// --- processOrder ---

func TestProcessOrder_OK_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockWorkerRepository(ctrl)

	mockRepo.EXPECT().
		UpdateOrderStatus(gomock.Any(), "111", "PROCESSED", 500.0).
		Return(nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := AccrualResponse{Order: "111", Status: "PROCESSED", Accrual: 500.0}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	w := NewAccrualWorker(mockRepo, server.URL, testLogger)
	w.processOrder(context.Background(), "111")
}

func TestProcessOrder_OK_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockWorkerRepository(ctrl)

	mockRepo.EXPECT().
		UpdateOrderStatus(gomock.Any(), "111", "PROCESSED", 500.0).
		Return(errors.New("db error"))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := AccrualResponse{Order: "111", Status: "PROCESSED", Accrual: 500.0}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	w := NewAccrualWorker(mockRepo, server.URL, testLogger)
	w.processOrder(context.Background(), "111")
}

func TestProcessOrder_OK_DecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not-json"))
	}))
	defer server.Close()

	w := NewAccrualWorker(nil, server.URL, testLogger)
	w.processOrder(context.Background(), "111")
}

func TestProcessOrder_NoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	w := NewAccrualWorker(nil, server.URL, testLogger)
	w.processOrder(context.Background(), "111")
}

func TestProcessOrder_TooManyRequests_WithRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	w := NewAccrualWorker(nil, server.URL, testLogger)
	w.processOrder(context.Background(), "111")
}

func TestProcessOrder_TooManyRequests_InvalidRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "abc")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	w := NewAccrualWorker(nil, server.URL, testLogger)
	w.processOrder(context.Background(), "111")
}

func TestProcessOrder_UnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	w := NewAccrualWorker(nil, server.URL, testLogger)
	w.processOrder(context.Background(), "111")
}

func TestProcessOrder_HTTPClientError(t *testing.T) {
	// Сервер сразу закрыт — запрос упадёт
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	w := NewAccrualWorker(nil, server.URL, testLogger)
	w.processOrder(context.Background(), "111")
}

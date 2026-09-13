package service

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
)

func TestAccrualClient_GetOrderInfo_OK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/orders/123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"order":"123","status":"PROCESSED","accrual":42.5}`))
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)
	resp, err := client.GetOrderInfo("123")
	if err != nil {
		t.Fatalf("GetOrderInfo() error = %v", err)
	}
	if resp.Order != "123" || resp.Status != domain.AccrualStatusProcessed {
		t.Errorf("unexpected response: %+v", resp)
	}
	if resp.Accrual == nil || *resp.Accrual != 42.5 {
		t.Errorf("resp.Accrual = %v, want 42.5", resp.Accrual)
	}
}

func TestAccrualClient_GetOrderInfo_NoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)
	resp, err := client.GetOrderInfo("123")
	if err != nil {
		t.Fatalf("GetOrderInfo() error = %v", err)
	}
	if resp.Status != domain.AccrualStatusRegistered {
		t.Errorf("resp.Status = %q, want REGISTERED", resp.Status)
	}
}

func TestAccrualClient_GetOrderInfo_TooManyRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)

	_, err := client.GetOrderInfo("123")

	var rateLimitErr *domain.RateLimitError
	if !errors.As(err, &rateLimitErr) {
		t.Fatalf("GetOrderInfo() error = %v, want *RateLimitError", err)
	}
	if rateLimitErr.RetryAfter != 60*time.Second {
		t.Errorf("RetryAfter = %v, want 60s", rateLimitErr.RetryAfter)
	}
}

func TestAccrualClient_GetOrderInfo_UnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)
	if _, err := client.GetOrderInfo("123"); err == nil {
		t.Error("GetOrderInfo() expected error, got nil")
	}
}

func TestAccrualClient_GetOrderInfo_BadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)
	if _, err := client.GetOrderInfo("123"); err == nil {
		t.Error("GetOrderInfo() expected error, got nil")
	}
}

func TestAccrualClient_GetOrderInfo_ConnectionError(t *testing.T) {
	client := NewAccrualClient("http://127.0.0.1:1")
	if _, err := client.GetOrderInfo("123"); err == nil {
		t.Error("GetOrderInfo() expected error, got nil")
	}
}

func TestAccrualClient_GetOrderInfo_RetryOn5xx(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// На третьей попытке — успех
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"order":"123","status":"PROCESSED","accrual":500}`))
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)

	resp, err := client.GetOrderInfo("123")
	requireNoError(t, err)
	if resp.Status != domain.AccrualStatusProcessed {
		t.Errorf("status = %s, want PROCESSED", resp.Status)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestAccrualClient_GetOrderInfo_ExhaustsRetries(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)

	_, err := client.GetOrderInfo("123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if attempts != maxRetries+1 {
		t.Errorf("attempts = %d, want %d", attempts, maxRetries+1)
	}
}

func TestAccrualClient_GetOrderInfo_NoRetryOn4xx(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)

	_, err := client.GetOrderInfo("123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1 (no retry on 4xx)", attempts)
	}
}

package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPClient_GetOrder_200(t *testing.T) {
	accrual := 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(OrderResponse{
			Order:   "12345678903",
			Status:  "PROCESSED",
			Accrual: &accrual,
		})
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, nil)
	resp, err := client.GetOrder(context.Background(), "12345678903")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Order != "12345678903" {
		t.Errorf("Order: got %q", resp.Order)
	}
	if resp.Status != "PROCESSED" {
		t.Errorf("Status: got %q", resp.Status)
	}
	if resp.Accrual == nil || *resp.Accrual != 500 {
		t.Errorf("Accrual: got %v", resp.Accrual)
	}
}

func TestHTTPClient_GetOrder_200_NoAccrual(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(OrderResponse{
			Order:   "456",
			Status:  "INVALID",
			Accrual: nil,
		})
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, nil)
	resp, err := client.GetOrder(context.Background(), "456")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if resp.Accrual != nil {
		t.Errorf("Accrual: expected nil, got %v", resp.Accrual)
	}
	if resp.Status != "INVALID" {
		t.Errorf("Status: got %q", resp.Status)
	}
}

func TestHTTPClient_GetOrder_204(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, nil)
	resp, err := client.GetOrder(context.Background(), "999")
	if err == nil {
		t.Fatal("expected error")
	}
	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}
	var notReg *ErrOrderNotRegistered
	if !errors.As(err, &notReg) {
		t.Fatalf("expected *ErrOrderNotRegistered, got %T: %v", err, err)
	}
}

func TestHTTPClient_GetOrder_429(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, nil)
	resp, err := client.GetOrder(context.Background(), "111")
	if err == nil {
		t.Fatal("expected error")
	}
	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}
	var rateLimit *ErrRateLimit
	if !errors.As(err, &rateLimit) {
		t.Fatalf("expected *ErrRateLimit, got %T: %v", err, err)
	}
}

func TestHTTPClient_GetOrder_500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, nil)
	resp, err := client.GetOrder(context.Background(), "333")
	if err == nil {
		t.Fatal("expected error")
	}
	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}
	if !errors.Is(err, ErrServerError) {
		t.Fatalf("expected errors.Is(ErrServerError), got %T: %v", err, err)
	}
	var ctxErr *ErrServerErrorContext
	if !errors.As(err, &ctxErr) {
		t.Fatalf("expected *ErrServerErrorContext, got %T", err)
	}
}

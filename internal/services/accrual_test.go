package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccrualService(t *testing.T) {
	baseURL := "http://localhost:8080"
	service := NewAccrualService(baseURL)

	require.NotNil(t, service)
	assert.Equal(t, baseURL, service.baseURL)
	require.NotNil(t, service.client)
}

func TestNewAccrualServiceWithRetry(t *testing.T) {
	baseURL := "http://localhost:8080"
	maxRetries := 5
	baseDelay := 200 * time.Millisecond
	maxDelay := 10 * time.Second

	service := NewAccrualServiceWithRetry(baseURL, maxRetries, baseDelay, maxDelay)

	require.NotNil(t, service)
	assert.Equal(t, baseURL, service.baseURL)
	require.NotNil(t, service.client)
}

func TestAccrualService_GetOrderInfo_Success(t *testing.T) {
	// Тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/orders/12345678903", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"order": "12345678903",
			"status": "PROCESSED",
			"accrual": 500
		}`))
	}))
	defer server.Close()

	service := NewAccrualService(server.URL)
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "12345678903", result.Order)
	assert.Equal(t, "PROCESSED", result.Status)
	require.NotNil(t, result.Accrual)
	assert.Equal(t, 500.0, *result.Accrual)
}

func TestAccrualService_GetOrderInfo_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	service := NewAccrualService(server.URL)
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestAccrualService_GetOrderInfo_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("No more than N requests per minute allowed"))
	}))
	defer server.Close()

	service := NewAccrualService(server.URL)
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.Error(t, err)
	assert.Equal(t, "rate limit exceeded", err.Error())
	assert.Nil(t, result)
}

func TestAccrualService_GetOrderInfo_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	service := NewAccrualService(server.URL)
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "giving up after")
	assert.Nil(t, result)
}

func TestAccrualService_GetOrderInfo_NetworkError(t *testing.T) {
	// Используем несуществующий URL для симуляции сетевой ошибки
	service := NewAccrualService("http://localhost:99999")
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAccrualService_GetOrderInfo_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	service := NewAccrualService(server.URL)
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAccrualService_GetOrderInfo_RetrySuccess(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// Первые два запроса возвращают ошибку
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			// Третий запрос успешен
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"order": "12345678903",
				"status": "PROCESSED",
				"accrual": 500
			}`))
		}
	}))
	defer server.Close()

	// Создаем сервис с быстрыми retry
	service := NewAccrualServiceWithRetry(server.URL, 3, 10*time.Millisecond, 100*time.Millisecond)
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "12345678903", result.Order)
	assert.Equal(t, "PROCESSED", result.Status)
	assert.Equal(t, 3, attempts) // Проверяем, что было 3 попытки
}

func TestAccrualService_GetOrderInfo_RetryExhausted(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		// Всегда возвращаем ошибку
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Создаем сервис с быстрыми retry
	service := NewAccrualServiceWithRetry(server.URL, 2, 10*time.Millisecond, 100*time.Millisecond)
	ctx := context.Background()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "giving up after")
	assert.Equal(t, 3, attempts) // Проверяем, что было 3 попытки (включая первую)
}

func TestAccrualService_GetOrderInfo_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Долгая задержка для симуляции медленного ответа
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewAccrualService(server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	result, err := service.GetOrderInfo(ctx, "12345678903")

	assert.Error(t, err)
	assert.Nil(t, result)
	// Проверяем, что ошибка связана с контекстом
	assert.Contains(t, err.Error(), "context")
}

package orderclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetOrderInfo(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/orders/123":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(OrderResponse{
				Order:   "123",
				Status:  StatusProcessed,
				Accrual: floatPtr(500.0),
			})

		case "/api/orders/456":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(OrderResponse{
				Order:  "456",
				Status: StatusProcessing,
			})

		case "/api/orders/notfound":
			w.WriteHeader(http.StatusNoContent)

		case "/api/orders/ratelimit":
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusTooManyRequests)

		case "/api/orders/servererror":
			w.WriteHeader(http.StatusInternalServerError)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Создаем клиент
	client := NewWithDefaults(
		WithBaseURL(server.URL),
		WithTimeout(5*time.Second),
	)

	ctx := context.Background()

	t.Run("success with accrual", func(t *testing.T) {
		resp, err := client.GetOrderInfo(ctx, "123")
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, "123", resp.Order)
		assert.Equal(t, StatusProcessed, resp.Status)
		assert.NotNil(t, resp.Accrual)
		assert.Equal(t, 500.0, *resp.Accrual)
		assert.True(t, resp.HasAccrual())
	})

	t.Run("success without accrual", func(t *testing.T) {
		resp, err := client.GetOrderInfo(ctx, "456")
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, "456", resp.Order)
		assert.Equal(t, StatusProcessing, resp.Status)
		assert.Nil(t, resp.Accrual)
		assert.False(t, resp.HasAccrual())
	})

	t.Run("order not found", func(t *testing.T) {
		resp, err := client.GetOrderInfo(ctx, "notfound")
		require.Error(t, err)
		require.Nil(t, resp)

		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, ErrorTypeNotFound, apiErr.Type)
		assert.Equal(t, http.StatusNoContent, apiErr.StatusCode)
		assert.True(t, apiErr.IsNotFound())
	})

	t.Run("rate limit", func(t *testing.T) {
		client.config.MaxRetries = 0 // Отключаем повторные попытки для теста
		resp, err := client.GetOrderInfo(ctx, "ratelimit")
		require.Error(t, err)
		require.Nil(t, resp)

		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, ErrorTypeRateLimit, apiErr.Type)
		assert.Equal(t, http.StatusTooManyRequests, apiErr.StatusCode)
		assert.Equal(t, "5", apiErr.RetryAfter)
		assert.True(t, apiErr.IsRateLimit())
		assert.True(t, apiErr.Retryable())
	})

	t.Run("server error", func(t *testing.T) {
		client.config.MaxRetries = 0
		resp, err := client.GetOrderInfo(ctx, "servererror")
		require.Error(t, err)
		require.Nil(t, resp)

		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, ErrorTypeServer, apiErr.Type)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.True(t, apiErr.Retryable())
	})
}

func floatPtr(f float64) *float64 {
	return &f
}

func TestStatusValidation(t *testing.T) {
	assert.True(t, StatusRegistered.IsValid())
	assert.True(t, StatusInvalid.IsValid())
	assert.True(t, StatusProcessing.IsValid())
	assert.True(t, StatusProcessed.IsValid())
	assert.False(t, OrderStatus("INVALID_STATUS").IsValid())
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"5", 5 * time.Second, false},
		{"10", 10 * time.Second, false},
		{"", 0, true},
		{"invalid", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := parseRetryAfter(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

package clients_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/KalessinD/gophermart/internal/clients"
	"github.com/KalessinD/gophermart/internal/models"
	"github.com/go-chi/chi/middleware"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockStep содержит данные для создания ответа, но не сам ответ (борьба со слепым линтером).
type mockStep struct {
	status  int
	body    []byte
	headers map[string]string
	err     error
}

// MockTransport реализует http.RoundTripper.
type MockTransport struct {
	steps     []mockStep
	callIndex int
}

func (m *MockTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	if m.callIndex >= len(m.steps) {
		return nil, errors.New("mock called more times than expected")
	}
	step := m.steps[m.callIndex]
	m.callIndex++

	if step.err != nil {
		return nil, step.err
	}

	// Создаем заголовки
	header := make(http.Header)
	for k, v := range step.headers {
		header.Set(k, v)
	}

	return &http.Response{
		StatusCode: step.status,
		Body:       io.NopCloser(bytes.NewReader(step.body)),
		Header:     header,
	}, nil
}

func TestAccrualClient_GetOrderAccrual(t *testing.T) {
	validOrderID := "12345678903"
	validResponse := models.AccrualResponse{
		Order:   validOrderID,
		Status:  "PROCESSED",
		Accrual: "500",
	}
	validBody, _ := json.Marshal(validResponse)

	tests := []struct {
		name       string
		orderID    string
		mockSteps  []mockStep
		wantErr    error
		wantResp   *models.AccrualResponse
		checkError func(t *testing.T, err error)
	}{
		{
			name:    "Success - Status OK",
			orderID: validOrderID,
			mockSteps: []mockStep{
				{status: http.StatusOK, body: validBody},
			},
			wantResp: &validResponse,
			wantErr:  nil,
		},
		{
			name:    "Error - Order Not Found (204)",
			orderID: "999",
			mockSteps: []mockStep{
				{status: http.StatusNoContent},
			},
			wantErr:  clients.ErrOrderNotFound,
			wantResp: nil,
		},
		{
			name:    "Error - Service Is Busy (429) with valid Retry-After",
			orderID: validOrderID,
			mockSteps: []mockStep{
				{
					status: http.StatusTooManyRequests,
					body:   []byte("No more than N requests"),
					headers: map[string]string{
						"Retry-After": "60",
					},
				},
			},
			wantErr: clients.ErrServiceIsBusy,
			checkError: func(t *testing.T, err error) {
				var busyErr *clients.ServiceIsBusyError
				require.True(t, errors.As(err, &busyErr), "error should be ServiceIsBusyError")
				require.Equal(t, 60*time.Second, busyErr.GetDelay(), "delay should match Retry-After header")
			},
		},
		{
			name:    "Error - Service Is Busy (429) with invalid Retry-After (use default)",
			orderID: validOrderID,
			mockSteps: []mockStep{
				{
					status: http.StatusTooManyRequests,
					body:   []byte("Error"),
					headers: map[string]string{
						"Retry-After": "invalid-value",
					},
				},
			},
			wantErr: clients.ErrServiceIsBusy,
			checkError: func(t *testing.T, err error) {
				var busyErr *clients.ServiceIsBusyError
				require.True(t, errors.As(err, &busyErr), "error should be ServiceIsBusyError")
				require.Equal(t, clients.DefaultDelay*time.Second, busyErr.GetDelay(), "delay should be default value")
			},
		},
		{
			name:    "Error - Internal Server Error (500)",
			orderID: validOrderID,
			mockSteps: []mockStep{
				{status: http.StatusInternalServerError},
			},
			wantErr: errors.New("unexpected status code: 500"),
		},
		{
			name:    "Retry Logic - Success after one failure",
			orderID: validOrderID,
			mockSteps: []mockStep{
				{err: errors.New("connection reset by peer")}, // Первая попытка - ошибка
				{status: http.StatusOK, body: validBody},      // Вторая - успех
			},
			wantResp: &validResponse,
			wantErr:  nil,
		},
		{
			name:    "Retry Logic - Fail after max attempts",
			orderID: validOrderID,
			mockSteps: []mockStep{
				{err: errors.New("timeout")},
				{err: errors.New("timeout")},
				{err: errors.New("timeout")},
			},
			wantErr: errors.New("request failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(t.Context(), middleware.RequestIDKey, "test-request-id-123")

			transport := &MockTransport{
				steps: tt.mockSteps,
			}

			httpClient := &http.Client{
				Transport: transport,
			}

			client := &clients.AccrualClient{
				BaseClient: httpClient,
				BaseURL:    "",
				Log:        zap.NewNop(),
			}

			resp, err := client.GetOrderAccrual(ctx, tt.orderID)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				require.Nil(t, resp)

				if tt.checkError != nil {
					tt.checkError(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tt.wantResp.Order, resp.Order)
				require.Equal(t, tt.wantResp.Status, resp.Status)
				require.Equal(t, tt.wantResp.Accrual, resp.Accrual)
			}
		})
	}
}

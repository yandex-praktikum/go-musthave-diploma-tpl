package worker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/AnxVit/go-musthave-diploma-tpl/internal/model"
)

func TestAccuralClient_GetResult(t *testing.T) {
	tests := []struct {
		name           string
		orderNumber    string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedResult *model.AccuralResponse
		expectedError  error
	}{
		{
			name:        "successful get result (200)",
			orderNumber: "12345678903",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/orders/12345678903", r.URL.Path)
				assert.Equal(t, "GET", r.Method)

				resp := model.AccuralResponse{
					Order:   "12345678903",
					Status:  "PROCESSED",
					Accrual: new(150.5),
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			expectedResult: &model.AccuralResponse{
				Order:   "12345678903",
				Status:  "PROCESSED",
				Accrual: new(150.5),
			},
			expectedError: nil,
		},
		{
			name:        "order not processed yet (204)",
			orderNumber: "12345678903",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:        "order not found",
			orderNumber: "99999999999",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error": "order not found"}`))
			},
			expectedResult: nil,
			expectedError:  &CodeError{code: http.StatusNotFound},
		},
		{
			name:        "server error",
			orderNumber: "12345678903",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedResult: nil,
			expectedError:  &CodeError{code: http.StatusInternalServerError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			ctx := context.Background()
			client := NewAccuralClient(ctx, server.URL)

			result, err := client.GetResult(tt.orderNumber)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if codeErr, ok := err.(*CodeError); ok {
					assert.Equal(t, tt.expectedError.(*CodeError).code, codeErr.code)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestAccuralClient_Register(t *testing.T) {
	tests := []struct {
		name           string
		orderNumber    string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedError  error
	}{
		{
			name:        "successful register",
			orderNumber: "12345678903",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/orders", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				var req model.AccuralRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, "12345678903", req.Order)

				w.WriteHeader(http.StatusOK)
			},
			expectedError: nil,
		},
		{
			name:        "order already registered (409)",
			orderNumber: "12345678903",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(`{"error": "order already registered"}`))
			},
			expectedError: nil,
		},
		{
			name:        "bad request (400)",
			orderNumber: "invalid",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "invalid order number"}`))
			},
			expectedError: &CodeError{code: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			ctx := context.Background()
			client := NewAccuralClient(ctx, server.URL)

			err := client.Register(tt.orderNumber)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if codeErr, ok := err.(*CodeError); ok {
					assert.Equal(t, tt.expectedError.(*CodeError).code, codeErr.code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

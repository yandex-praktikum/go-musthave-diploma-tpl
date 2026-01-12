package accrual

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetOrderInfo(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		retryAfter    string
		wantAccrual   *OrderAccrual
		wantErr       bool
		wantRateLimit bool
	}{
		{
			name:       "success processed order",
			statusCode: http.StatusOK,
			responseBody: `{
				"order": "12345678903",
				"status": "PROCESSED",
				"accrual": 500.5
			}`,
			wantAccrual: &OrderAccrual{
				Order:   "12345678903",
				Status:  StatusProcessed,
				Accrual: floatPtr(500.5),
			},
			wantErr: false,
		},
		{
			name:       "success processing order",
			statusCode: http.StatusOK,
			responseBody: `{
				"order": "9278923470",
				"status": "PROCESSING"
			}`,
			wantAccrual: &OrderAccrual{
				Order:   "9278923470",
				Status:  StatusProcessing,
				Accrual: nil,
			},
			wantErr: false,
		},
		{
			name:        "not registered",
			statusCode:  http.StatusNoContent,
			wantAccrual: nil,
			wantErr:     false,
		},
		{
			name:          "rate limit",
			statusCode:    http.StatusTooManyRequests,
			retryAfter:    "60",
			wantAccrual:   nil,
			wantErr:       true,
			wantRateLimit: true,
		},
		{
			name:        "server error",
			statusCode:  http.StatusInternalServerError,
			wantAccrual: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.retryAfter != "" {
					w.Header().Set("Retry-After", tt.retryAfter)
				}
				w.WriteHeader(tt.statusCode)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			client, err := New(server.URL)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			ctx := context.Background()
			got, err := client.GetOrderInfo(ctx, "12345678903")

			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrderInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantRateLimit {
				if _, ok := err.(*RateLimitError); !ok {
					t.Errorf("GetOrderInfo() error = %v, want RateLimitError", err)
					return
				}
				rateLimitErr := err.(*RateLimitError)
				if rateLimitErr.RetryAfter != 60*time.Second {
					t.Errorf("GetOrderInfo() RetryAfter = %v, want %v", rateLimitErr.RetryAfter, 60*time.Second)
				}
				return
			}

			if !tt.wantErr {
				if got == nil && tt.wantAccrual != nil {
					t.Errorf("GetOrderInfo() = nil, want %+v", tt.wantAccrual)
					return
				}
				if got != nil && tt.wantAccrual == nil {
					t.Errorf("GetOrderInfo() = %+v, want nil", got)
					return
				}
				if got != nil && tt.wantAccrual != nil {
					if got.Order != tt.wantAccrual.Order {
						t.Errorf("GetOrderInfo().Order = %v, want %v", got.Order, tt.wantAccrual.Order)
					}
					if got.Status != tt.wantAccrual.Status {
						t.Errorf("GetOrderInfo().Status = %v, want %v", got.Status, tt.wantAccrual.Status)
					}
					if got.Accrual != nil && tt.wantAccrual.Accrual != nil {
						if *got.Accrual != *tt.wantAccrual.Accrual {
							t.Errorf("GetOrderInfo().Accrual = %v, want %v", *got.Accrual, *tt.wantAccrual.Accrual)
						}
					} else if got.Accrual != tt.wantAccrual.Accrual {
						t.Errorf("GetOrderInfo().Accrual = %v, want %v", got.Accrual, tt.wantAccrual.Accrual)
					}
				}
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		wantErr bool
	}{
		{
			name:    "valid URL",
			rawURL:  "http://localhost:8080",
			wantErr: false,
		},
		{
			name:    "invalid URL",
			rawURL:  "://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("New() = nil, want non-nil Client")
			}
		})
	}
}

func TestRateLimitError_Error(t *testing.T) {
	err := &RateLimitError{RetryAfter: 60 * time.Second}
	got := err.Error()
	if got == "" {
		t.Error("RateLimitError.Error() returned empty string")
	}
	if !contains(got, "rate limit") {
		t.Errorf("RateLimitError.Error() = %v, want to contain 'rate limit'", got)
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package models_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/KalessinD/gophermart/internal/models"
)

func TestNewOrder(t *testing.T) {
	id := "123456789"
	userID := "user_1"
	status := models.OrderNewStatus

	order, err := models.NewOrder(id, userID, status)
	if err != nil {
		t.Fatalf("NewOrder() returned unexpected error: %v", err)
	}
	if order == nil {
		t.Fatal("NewOrder() returned nil order")
	}
	if order.ID != id {
		t.Errorf("Expected ID %s, got %s", id, order.ID)
	}
	if order.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, order.UserID)
	}
	if order.Status != status {
		t.Errorf("Expected Status %s, got %s", status, order.Status)
	}
}

func TestOrder_Validate(t *testing.T) {
	tests := []struct {
		name    string
		order   models.Order
		wantErr bool
		errMsg  string // Подстрока, ожидаемая в ошибке
	}{
		{
			name:    "valid order NEW",
			order:   models.Order{ID: "1", UserID: "u1", Status: models.OrderNewStatus},
			wantErr: false,
		},
		{
			name:    "valid order PROCESSED with accrual",
			order:   models.Order{ID: "1", UserID: "u1", Status: models.OrderProcessedStatus, Accrual: 500},
			wantErr: false,
		},
		{
			name:    "empty UserID",
			order:   models.Order{ID: "1", UserID: "", Status: models.OrderNewStatus},
			wantErr: true,
			errMsg:  "UserID can't be empty",
		},
		{
			name:    "empty ID",
			order:   models.Order{ID: "", UserID: "u1", Status: models.OrderNewStatus},
			wantErr: true,
			errMsg:  "ID can't be equal to zero",
		},
		{
			name:    "invalid status",
			order:   models.Order{ID: "1", UserID: "u1", Status: "UNKNOWN"},
			wantErr: true,
			errMsg:  "wrong status value",
		},
		{
			name:    "negative accrual",
			order:   models.Order{ID: "1", UserID: "u1", Status: models.OrderNewStatus, Accrual: -10},
			wantErr: true,
			errMsg:  "accrual must be a positive or zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestOrder_ToJSON(t *testing.T) {
	// Фиксируем время для предсказуемости теста
	now := time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		order    models.Order
		wantJSON string // Ожидаемая подстрока или полный JSON
	}{
		{
			name: "accrual is 0 (should be omitted)",
			order: models.Order{
				ID:         "123",
				UserID:     "u1", // Должен быть скрыт
				Status:     models.OrderNewStatus,
				Accrual:    0, // Должен быть скрыт
				UploadedAt: now,
			},
			wantJSON: `{"number":"123","status":"NEW","uploaded":"2023-10-01T12:00:00Z"}`,
		},
		{
			name: "accrual is positive (should be present)",
			order: models.Order{
				ID:         "456",
				UserID:     "u1",
				Status:     models.OrderProcessedStatus,
				Accrual:    500,
				UploadedAt: now,
			},
			wantJSON: `{"number":"456","status":"PROCESSED","accrual":5.00,"uploaded":"2023-10-01T12:00:00Z"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.order.ToJSON()
			if err != nil {
				t.Fatalf("ToJSON() returned unexpected error: %v", err)
			}

			gotString := string(data)
			if gotString != tt.wantJSON {
				t.Errorf("ToJSON() = %s, want %s", gotString, tt.wantJSON)
			}

			// убеждаемся, что UserID нет в JSON
			if strings.Contains(gotString, "UserID") {
				t.Error("ToJSON() result should not contain UserID")
			}
			if strings.Contains(gotString, "UpdatedAt") {
				t.Error("ToJSON() result should not contain UpdatedAt")
			}
		})
	}
}

func TestOrder_SetStatus(t *testing.T) {
	order, err := models.NewOrder("123", "user1", models.OrderNewStatus)
	if err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	tests := []struct {
		name        string
		newStatus   string
		wantErr     bool
		expectedErr error
	}{
		{
			name:      "set PROCESSING status",
			newStatus: models.OrderInProcessStatus,
			wantErr:   false,
		},
		{
			name:      "set INVALID status",
			newStatus: models.OrderInvalidStatus,
			wantErr:   false,
		},
		{
			name:      "set PROCESSED status",
			newStatus: models.OrderProcessedStatus,
			wantErr:   false,
		},
		{
			name:      "set NEW status (same)",
			newStatus: models.OrderNewStatus,
			wantErr:   false,
		},
		{
			name:        "invalid status",
			newStatus:   "UNKNOWN",
			wantErr:     true,
			expectedErr: models.ErrORderWrongStatus,
		},
		{
			name:        "empty status",
			newStatus:   "",
			wantErr:     true,
			expectedErr: models.ErrORderWrongStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := order.SetStatus(tt.newStatus)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("SetStatus() error = %v, expectedErr %v", err, tt.expectedErr)
				}
				// Проверяем, что статус не изменился при ошибке
				if order.Status == tt.newStatus {
					t.Error("SetStatus() should not change status on error")
				}
			} else if order.Status != tt.newStatus {
				t.Errorf("SetStatus() status = %v, want %v", order.Status, tt.newStatus)
			}
		})
	}
}

func TestOrder_SetAccrual(t *testing.T) {
	order, err := models.NewOrder("123", "user1", models.OrderNewStatus)
	if err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	tests := []struct {
		name            string
		accrualStr      string
		wantErr         bool
		expectedAccrual models.Accrual
	}{
		{
			name:            "integer value",
			accrualStr:      "100",
			wantErr:         false,
			expectedAccrual: 10000, // 100.00 -> 10000 копеек
		},
		{
			name:            "value with one decimal",
			accrualStr:      "100.5",
			wantErr:         false,
			expectedAccrual: 10050, // 100.50 -> 10050 копеек
		},
		{
			name:            "value with two decimals",
			accrualStr:      "100.50",
			wantErr:         false,
			expectedAccrual: 10050,
		},
		{
			name:            "value with three decimals (truncated)",
			accrualStr:      "100.555",
			wantErr:         false,
			expectedAccrual: 10055, // обрезается до 2 знаков
		},
		{
			name:            "zero value",
			accrualStr:      "0",
			wantErr:         false,
			expectedAccrual: 0,
		},
		{
			name:            "small fractional value",
			accrualStr:      "0.05",
			wantErr:         false,
			expectedAccrual: 5,
		},
		{
			name:            "small value with one decimal",
			accrualStr:      "0.5",
			wantErr:         false,
			expectedAccrual: 50,
		},
		{
			name:       "invalid format - two dots",
			accrualStr: "100.1.1",
			wantErr:    true,
		},
		{
			name:       "invalid format - letters",
			accrualStr: "abc",
			wantErr:    true,
		},
		{
			name:       "invalid format - empty string",
			accrualStr: "",
			wantErr:    true,
		},
		{
			name:       "invalid format - only dot",
			accrualStr: ".",
			wantErr:    true,
		},
		{
			name:       "invalid format - letters in fractional part",
			accrualStr: "100.ab",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сбрасываем перед каждым тестом
			order.Accrual = 0

			err := order.SetAccrual(tt.accrualStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetAccrual() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && order.Accrual != tt.expectedAccrual {
				t.Errorf("SetAccrual() accrual = %v, want %v", order.Accrual, tt.expectedAccrual)
			}
		})
	}
}

func TestAccrual_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		accrual  models.Accrual
		wantJSON string
		wantErr  bool
	}{
		{
			name:     "zero value",
			accrual:  0,
			wantJSON: "0",
		},
		{
			name:     "5 rubles",
			accrual:  500,
			wantJSON: "5.00",
		},
		{
			name:     "50 kopecks",
			accrual:  50,
			wantJSON: "0.50",
		},
		{
			name:     "5 kopecks",
			accrual:  5,
			wantJSON: "0.05",
		},
		{
			name:     "12 rubles 34 kopecks",
			accrual:  1234,
			wantJSON: "12.34",
		},
		{
			name:     "1 ruble",
			accrual:  100,
			wantJSON: "1.00",
		},
		{
			name:     "1 kopeck",
			accrual:  1,
			wantJSON: "0.01",
		},
		{
			name:     "large value - 1000 rubles",
			accrual:  100000,
			wantJSON: "1000.00",
		},
		{
			name:     "negative value - 5 rubles",
			accrual:  -500,
			wantJSON: "-5.00",
		},
		{
			name:     "negative value - 50 kopecks",
			accrual:  -50,
			wantJSON: "-0.50",
		},
		{
			name:     "negative value - 5 kopecks",
			accrual:  -5,
			wantJSON: "-0.05",
		},
		{
			name:     "negative value - 12 rubles 34 kopecks",
			accrual:  -1234,
			wantJSON: "-12.34",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.accrual.MarshalJSON()

			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && string(data) != tt.wantJSON {
				t.Errorf("MarshalJSON() = %s, want %s", string(data), tt.wantJSON)
			}
		})
	}
}

func TestOrder_ToJSON_WithAccrual(t *testing.T) {
	now := time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		order        models.Order
		wantContains []string
		dontContains []string
	}{
		{
			name: "small accrual 0.01",
			order: models.Order{
				ID:         "123",
				UserID:     "u1",
				Status:     models.OrderProcessedStatus,
				Accrual:    1,
				UploadedAt: now,
			},
			wantContains: []string{`"accrual":0.01`},
			dontContains: []string{"UserID", "UpdatedAt"},
		},
		{
			name: "large accrual 1234.56",
			order: models.Order{
				ID:         "456",
				UserID:     "u1",
				Status:     models.OrderProcessedStatus,
				Accrual:    123456,
				UploadedAt: now,
			},
			wantContains: []string{`"accrual":1234.56`},
			dontContains: []string{"UserID", "UpdatedAt"},
		},
		{
			name: "accrual 0 should be omitted",
			order: models.Order{
				ID:         "789",
				UserID:     "u1",
				Status:     models.OrderNewStatus,
				Accrual:    0,
				UploadedAt: now,
			},
			wantContains: []string{`"number":"789"`, `"status":"NEW"`},
			dontContains: []string{"accrual", "UserID", "UpdatedAt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.order.ToJSON()
			if err != nil {
				t.Fatalf("ToJSON() returned unexpected error: %v", err)
			}

			gotString := string(data)

			for _, contain := range tt.wantContains {
				if !strings.Contains(gotString, contain) {
					t.Errorf("ToJSON() result should contain %q, got %q", contain, gotString)
				}
			}

			for _, dontContain := range tt.dontContains {
				if strings.Contains(gotString, dontContain) {
					t.Errorf("ToJSON() result should not contain %q, got %q", dontContain, gotString)
				}
			}
		})
	}
}

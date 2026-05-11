package models_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/KalessinD/gophermart/internal/models"
)

func TestWithdrawn_ToJSON(t *testing.T) {
	// Фиксируем время для предсказуемости теста
	now := time.Date(2023, 10, 5, 15, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		withdrawn    models.Withdrawal
		wantJSON     string
		wantContains []string
		dontContains []string
	}{
		{
			name: "valid withdrawal with sum",
			withdrawn: models.Withdrawal{
				ID:          "internal_id_123", // Должен быть скрыт
				UserID:      "user_1",          // Должен быть скрыт
				OrderID:     "order_abc",
				Sum:         500, // 5.00
				ProcessedAt: now,
			},
			wantJSON: `{"order":"order_abc","sum":5.00,"processed_at":"2023-10-05T15:30:00Z"}`,
		},
		{
			name: "small sum",
			withdrawn: models.Withdrawal{
				OrderID:     "order_small",
				Sum:         5, // 0.05
				ProcessedAt: now,
			},
			wantJSON: `{"order":"order_small","sum":0.05,"processed_at":"2023-10-05T15:30:00Z"}`,
		},
		{
			name: "zero sum",
			withdrawn: models.Withdrawal{
				OrderID:     "order_zero",
				Sum:         0,
				ProcessedAt: now,
			},
			wantJSON: `{"order":"order_zero","sum":0,"processed_at":"2023-10-05T15:30:00Z"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.withdrawn.ToJSON()
			if err != nil {
				t.Fatalf("ToJSON() returned unexpected error: %v", err)
			}

			gotString := string(data)

			// Проверка полного совпадения, если задано wantJSON
			if tt.wantJSON != "" && gotString != tt.wantJSON {
				t.Errorf("ToJSON() = %s, want %s", gotString, tt.wantJSON)
			}

			// Проверка, что скрытые поля не попали в JSON
			if strings.Contains(gotString, `"ID"`) {
				t.Error("ToJSON() result should not contain ID field")
			}
			if strings.Contains(gotString, `"UserID"`) {
				t.Error("ToJSON() result should not contain UserID field")
			}

			// Проверка, что OrderID отображается как "order"
			if strings.Contains(gotString, `"OrderID"`) {
				t.Error("ToJSON() result should not contain OrderID field (should be 'order')")
			}
		})
	}
}

func TestWithdrawnList_MarshalJSON(t *testing.T) {
	now := time.Date(2023, 10, 5, 15, 30, 0, 0, time.UTC)

	list := models.WithdrawalsList{
		{
			OrderID:     "order1",
			Sum:         100, // 1.00
			ProcessedAt: now,
		},
		{
			OrderID:     "order2",
			Sum:         2050, // 20.50
			ProcessedAt: now,
		},
	}

	data, err := json.Marshal(list)
	if err != nil {
		t.Fatalf("Marshal() returned unexpected error: %v", err)
	}

	gotString := string(data)

	// Проверяем, что это массив
	if !strings.HasPrefix(gotString, "[") || !strings.HasSuffix(gotString, "]") {
		t.Errorf("Result should be a JSON array, got %s", gotString)
	}

	// Проверяем наличие элементов
	if !strings.Contains(gotString, `"order":"order1"`) {
		t.Errorf("Result should contain order1, got %s", gotString)
	}
	if !strings.Contains(gotString, `"sum":1.00`) {
		t.Errorf("Result should contain sum 1.00, got %s", gotString)
	}
	if !strings.Contains(gotString, `"sum":20.50`) {
		t.Errorf("Result should contain sum 20.50, got %s", gotString)
	}
}

func TestErrWithdrawnNotFound(t *testing.T) {
	err := models.ErrWithdrawnNotFound
	if err == nil {
		t.Error("ErrWithdrawnNotFound should not be nil")
	}
	if err.Error() != "withdrawns not found" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

package models_test

import (
	"strings"
	"testing"

	"github.com/KalessinD/gophermart/internal/models"
)

func TestNewBalanceItem(t *testing.T) {
	tests := []struct {
		name          string
		balance       models.Accrual
		withdrawn     models.Accrual
		wantBalance   models.Accrual
		wantWithdrawn models.Accrual
	}{
		{
			name:          "standard values",
			balance:       10000, // 100.00
			withdrawn:     500,   // 5.00
			wantBalance:   10000,
			wantWithdrawn: 500,
		},
		{
			name:          "zero values",
			balance:       0,
			withdrawn:     0,
			wantBalance:   0,
			wantWithdrawn: 0,
		},
		{
			name:          "small fractional values",
			balance:       1, // 0.01
			withdrawn:     2, // 0.02
			wantBalance:   1,
			wantWithdrawn: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := models.NewBalanceItem(tt.balance, tt.withdrawn)

			if got == nil {
				t.Fatal("NewBalanceItem() returned nil")
			}

			if got.Balance != tt.wantBalance {
				t.Errorf("Balance field = %v, want %v", got.Balance, tt.wantBalance)
			}

			if got.Withdrawn != tt.wantWithdrawn {
				t.Errorf("Withdrawn field = %v, want %v", got.Withdrawn, tt.wantWithdrawn)
			}
		})
	}
}

func TestBalance_ToJSON(t *testing.T) {
	tests := []struct {
		name         string
		balance      models.Balance
		wantJSON     string
		wantContains []string
		dontContains []string
	}{
		{
			name: "valid balance with values",
			balance: models.Balance{
				Balance:   500,  // 5.00
				Withdrawn: 1000, // 10.00
			},
			wantJSON: `{"current":5.00,"withdrawn":10.00}`,
		},
		{
			name: "zero balance",
			balance: models.Balance{
				Balance:   0,
				Withdrawn: 0,
			},
			wantJSON: `{"current":0,"withdrawn":0}`,
		},
		{
			name: "small fractional values",
			balance: models.Balance{
				Balance:   5,   // 0.05
				Withdrawn: 105, // 1.05
			},
			wantJSON: `{"current":0.05,"withdrawn":1.05}`,
		},
		{
			name: "check field names",
			balance: models.Balance{
				Balance:   100,
				Withdrawn: 200,
			},
			wantContains: []string{`"current"`, `"withdrawn"`},
			dontContains: []string{`"Balance"`, `"Withdrawn"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.balance.ToJSON()
			if err != nil {
				t.Fatalf("ToJSON() returned unexpected error: %v", err)
			}

			gotString := string(data)

			// Проверка полного совпадения
			if tt.wantJSON != "" && gotString != tt.wantJSON {
				t.Errorf("ToJSON() = %s, want %s", gotString, tt.wantJSON)
			}

			// Проверка наличия подстрок
			for _, contain := range tt.wantContains {
				if !strings.Contains(gotString, contain) {
					t.Errorf("ToJSON() result should contain %q, got %q", contain, gotString)
				}
			}

			// Проверка отсутствия подстрок
			for _, dontContain := range tt.dontContains {
				if strings.Contains(gotString, dontContain) {
					t.Errorf("ToJSON() result should not contain %q, got %q", dontContain, gotString)
				}
			}
		})
	}
}

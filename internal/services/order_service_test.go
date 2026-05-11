package services

import (
	"testing"
)

func TestValidateOrderNumber(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"valid luhn 1", "12345678903", true}, // Valid Luhn
		{"valid luhn 2", "9278923470", true},  // Valid Luhn
		{"valid luhn 3", "346436439", true},   // Valid Luhn
		{"valid luhn 4", "2377225624", true},  // Valid Luhn
		{"empty string", "", false},
		{"non-digit chars", "12345abc", false},
		{"only letters", "abcdef", false},
		{"single digit", "0", true},
		{"invalid luhn", "1234567890", false},
		{"spaces", "123 456", false},
		{"negative", "-123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateOrderNumber(tt.number); got != tt.want {
				t.Errorf("ValidateOrderNumber(%q) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

func TestLuhnCheck(t *testing.T) {
	tests := []struct {
		number string
		want   bool
	}{
		// Valid Luhn numbers
		{"4532015112830366", true},
		{"6011514433546201", true},
		{"374245455400126", true},
		{"5425233430109903", true},
		// Invalid Luhn numbers
		{"4532015112830367", false},
		{"1234567890123456", false},
	}

	for _, tt := range tests {
		t.Run(tt.number, func(t *testing.T) {
			if got := luhnCheck(tt.number); got != tt.want {
				t.Errorf("luhnCheck(%q) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

func TestParseOrderNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid number", "12345678903", "12345678903", false},
		{"with spaces", "  12345678903  ", "12345678903", false},
		{"empty string", "", "", true},
		{"only spaces", "   ", "", true},
		{"with newline", "12345678903\n", "12345678903", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOrderNumber(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOrderNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseOrderNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntToPointerInt64(t *testing.T) {
	val := 42
	ptr := IntToPointerInt64(val)
	if ptr == nil {
		t.Fatal("IntToPointerInt64() returned nil")
	}
	if *ptr != 42 {
		t.Errorf("IntToPointerInt64() = %v, want 42", *ptr)
	}
}

func TestParseAccrual(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    int64
		wantNil bool
		wantErr bool
	}{
		{"int", int(100), 100, false, false},
		{"int64", int64(200), 200, false, false},
		{"float64", float64(300.5), 300, false, false},
		{"string valid", "400", 400, false, false},
		{"nil", nil, 0, true, false},
		{"string invalid", "abc", 0, true, true},
		{"invalid type", true, 0, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAccrual(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAccrual() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNil {
				if got != nil {
					t.Errorf("ParseAccrual() = %v, want nil", got)
				}
			} else if got == nil {
				t.Error("ParseAccrual() returned nil")
			} else if *got != tt.want {
				t.Errorf("ParseAccrual() = %v, want %v", *got, tt.want)
			}
		})
	}
}

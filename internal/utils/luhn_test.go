package utils

import "testing"

func TestValidLuhn(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"valid number", "4561261212345467", true},
		{"valid number 2", "79927398713", true},
		{"invalid checksum", "4561261212345468", false},
		{"empty string", "", false},
		{"contains letters", "4561abc2345467", false},
		{"contains symbol", "4561-234", false},
		{"single zero", "0", true},
		{"single one", "1", false},
		{"spaces", "4561 2345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidLuhn(tt.number); got != tt.want {
				t.Errorf("ValidLuhn(%q) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

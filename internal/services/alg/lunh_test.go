package alg_test

import (
	"testing"

	"github.com/KalessinD/gophermart/internal/services/alg"
)

func TestIsValidLuhn(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{
			name:     "Valid simple number",
			number:   "79927398713",
			expected: true,
		},
		{
			name:     "Valid number with spaces",
			number:   "7992 7398 713",
			expected: true,
		},
		{
			name:     "Valid number with dashes",
			number:   "7992-7398-713",
			expected: true,
		},
		{
			name:     "Valid Visa test number",
			number:   "4111111111111111",
			expected: true,
		},

		// --- Некорректные номера ---
		{
			name:     "Invalid checksum (last digit wrong)",
			number:   "79927398710", // Изменили последнюю цифру с 3 на 0
			expected: false,
		},
		{
			name:     "Invalid checksum (wrong digit in middle)",
			number:   "4111111111111112", // Изменена последняя цифра
			expected: false,
		},
		{
			name:     "Contains letters",
			number:   "7992abc98713",
			expected: false,
		},
		{
			name:     "Empty string",
			number:   "",
			expected: true,
		},
		{
			name:     "Short valid number",
			number:   "0",
			expected: true,
		},
		{
			name:     "Short invalid number",
			number:   "1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := alg.IsValidLuhn(tt.number)

			if result != tt.expected {
				t.Errorf("IsValidLuhn(%q) = %v, want %v", tt.number, result, tt.expected)
			}
		})
	}
}

package luhn

import (
	"gmart/internal/domain"
	"testing"
)

func TestIsValid(t *testing.T) {
	tests := []struct {
		name   string
		number domain.OrderNumber
		want   bool
	}{
		// Валидные номера
		{"valid short", "79927398713", true},
		{"valid standard", "49927398716", true},
		{"valid long", "1234567812345670", true},
		{"zero is valid", "0", true},

		// Невалидные номера (контрольная сумма не совпадает)
		{"invalid checksum", "79927398710", false},
		{"invalid simple", "123", false},
		{"all same digits invalid", "1111", false},

		// Некорректные входные данные
		{"empty string", "", true}, // сумма 0 % 10 == 0
		{"contains letters", "123a5", false},
		{"contains spaces", "123 5", false},
		{"contains symbols", "123-5", false},
		{"non-digit start", "a123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValid(tt.number); got != tt.want {
				t.Errorf("IsValid(%s) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

// Benchmark для проверки производительности (алгоритм не должен аллоцировать)
func BenchmarkIsValid(b *testing.B) {
	num := domain.OrderNumber("79927398713")
	for i := 0; i < b.N; i++ {
		IsValid(num)
	}
}

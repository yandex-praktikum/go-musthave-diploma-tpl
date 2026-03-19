package domain

import (
	"testing"
)

func TestAmount_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		m       Amount
		want    string
		wantErr bool
	}{
		{"zero", 0, "0.00", false},
		{"cents only", 5, "0.05", false},
		{"ten cents", 10, "0.10", false},
		{"rubles and cents", 12345, "123.45", false},
		{"large value", 18446744073709551615, "184467440737095516.15", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAmount_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Amount
		wantErr bool
	}{
		// Успешные кейсы
		{"empty", "", 0, false},
		{"null", "null", 0, false},
		{"integer", "100", 10000, false},
		{"float with 2 digits", "10.55", 1055, false},
		{"float with 1 digit", "10.5", 1050, false},
		{"quoted string", `"10.50"`, 1050, false},
		{"extra zero tolerance", "10.5000", 1050, false},
		{"leading zeros", "00010.50", 1050, false},
		{"spaces and quotes", `  "10.50"  `, 1050, false},
		{"zero value float", "0.00", 0, false},

		// Ошибки (покрытие веток)
		{"multiple dots", "10.50.10", 0, true},
		{"invalid character", "10a50", 0, true},
		{"invalid character in cents", "10.5a", 0, true},
		{"too much precision (non-zero)", "10.501", 0, true},
		{"too much precision (long)", "10.50000001", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m Amount
			err := m.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && m != tt.want {
				t.Errorf("UnmarshalJSON(%s) = %d, want %d", tt.input, m, tt.want)
			}
		})
	}
}

// Benchmark для проверки аллокаций
func BenchmarkAmount_MarshalJSON(b *testing.B) {
	m := Amount(123456789)
	for i := 0; i < b.N; i++ {
		_, _ = m.MarshalJSON()
	}
}

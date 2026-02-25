package luhn

import "testing"

func TestValid(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "valid number 79927398713",
			number: "79927398713",
			want:   true,
		},
		{
			name:   "valid number 4532015112830366",
			number: "4532015112830366",
			want:   true,
		},
		{
			name:   "valid number 6011514433546201",
			number: "6011514433546201",
			want:   true,
		},
		{
			name:   "valid number 49927398716",
			number: "49927398716",
			want:   true,
		},
		{
			name:   "valid single digit 0",
			number: "0",
			want:   true,
		},
		{
			name:   "invalid number 79927398710",
			number: "79927398710",
			want:   false,
		},
		{
			name:   "invalid number 1234567890",
			number: "1234567890",
			want:   false,
		},
		{
			name:   "invalid - contains letters",
			number: "1234a567890",
			want:   false,
		},
		{
			name:   "invalid - empty string",
			number: "",
			want:   false,
		},
		{
			name:   "invalid - only letters",
			number: "abcdef",
			want:   false,
		},
		{
			name:   "invalid - special characters",
			number: "1234-5678",
			want:   false,
		},
		{
			name:   "valid number 18",
			number: "18",
			want:   true,
		},
		{
			name:   "valid two digits",
			number: "59",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Valid(tt.number); got != tt.want {
				t.Errorf("Valid(%q) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

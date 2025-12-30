package server

import "testing"

func TestIsValidOrderNumber(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "valid luhn number",
			number: "12345678903",
			want:   true,
		},
		{
			name:   "valid luhn number 2",
			number: "9278923470",
			want:   true,
		},
		{
			name:   "invalid luhn number",
			number: "12345678904",
			want:   false,
		},
		{
			name:   "empty string",
			number: "",
			want:   false,
		},
		{
			name:   "non-digits",
			number: "123abc456",
			want:   false,
		},
		{
			name:   "single digit",
			number: "5",
			want:   false,
		},
		{
			name:   "valid short number",
			number: "42",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidOrderNumber(tt.number)
			if got != tt.want {
				t.Errorf("isValidOrderNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

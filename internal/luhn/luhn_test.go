package luhn

import "testing"

func TestValid(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{name: "valid", number: "12345678903", want: true},
		{name: "invalid", number: "12345678904", want: false},
		{name: "empty", number: "", want: false},
		{name: "nondigits", number: "12345abc", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Valid(tt.number); got != tt.want {
				t.Fatalf("Valid(%q) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

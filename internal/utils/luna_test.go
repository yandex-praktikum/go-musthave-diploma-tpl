package utils

import "testing"

func TestIsLunaValid(t *testing.T) {
	tests := []struct {
		name           string
		order          string
		expentedResult bool
	}{
		{
			name:           "Successful order",
			order:          "22664155",
			expentedResult: true,
		},
		{
			name:  "Invalid order",
			order: "aaaa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := IsLunaValid(tt.order)

			if valid != tt.expentedResult {
				t.Errorf("expented result %t, got %t", tt.expentedResult, valid)
			}
		})
	}
}

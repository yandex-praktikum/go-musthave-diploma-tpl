package ptr

import (
	"testing"
)

func TestStrPtr(t *testing.T) {
	testCases := []struct {
		input    string
		expected *string
	}{
		{"Hello, World!", StrPtr("Hello, World!")},
		{"", StrPtr("")},
		{"123", StrPtr("123")},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := StrPtr(tc.input)

			if result == nil && tc.expected != nil {
				t.Errorf("Expected a non-nil result, but got nil.")
			} else if result != nil && tc.expected == nil {
				t.Errorf("Expected a nil result, but got a non-nil result.")
			} else if result != nil && *result != *tc.expected {
				t.Errorf("Expected %v, but got %v", *tc.expected, *result)
			}
		})
	}
}

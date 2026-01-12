package service

import (
	"errors"
	"testing"
)

func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "duplicate key error",
			err:  errors.New("duplicate key value violates unique constraint"),
			want: true,
		},
		{
			name: "no error",
			err:  nil,
			want: false,
		},
		{
			name: "other error",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "error with duplicate key in message",
			err:  errors.New("error: duplicate key value violates unique constraint \"users_login_key\""),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUniqueViolation(tt.err)
			if got != tt.want {
				t.Errorf("isUniqueViolation() = %v, want %v", got, tt.want)
			}
		})
	}
}

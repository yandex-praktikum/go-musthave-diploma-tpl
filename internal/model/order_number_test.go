package model

import (
	"errors"
	"testing"
)

const (
	invalidOrderNumber    = "123"
	validOrderNumber      = "4368910403255712"
	nonNumericOrderNumber = "GOPHER"
)

func TestValidateOrderNumber(t *testing.T) {
	type args struct {
		number string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name:    "empty → error",
			args:    args{number: ""},
			wantErr: ErrMissingOrderNumber,
		},
		{
			name:    "invalid number → error",
			args:    args{number: invalidOrderNumber},
			wantErr: ErrInvalidOrderNumber,
		},
		{
			name:    "valid number → ok",
			args:    args{number: validOrderNumber},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateOrderNumber(tt.args.number); !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateOrderNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_luhnCheck(t *testing.T) {
	type args struct {
		orderNumber string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty → false",
			args: args{orderNumber: ""},
			want: false,
		},
		{
			name: "invalid number → false",
			args: args{orderNumber: invalidOrderNumber},
			want: false,
		},
		{
			name: "non-numeric → false",
			args: args{orderNumber: nonNumericOrderNumber},
			want: false,
		},
		{
			name: "valid number → true",
			args: args{orderNumber: validOrderNumber},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := luhnCheck(tt.args.orderNumber); got != tt.want {
				t.Errorf("luhnCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

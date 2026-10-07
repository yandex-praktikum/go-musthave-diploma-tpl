package model

import (
	"errors"
	"testing"
)

func TestValidateWithdrawalAmount(t *testing.T) {
	type args struct {
		amount float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name:    "positive → ok",
			args:    args{amount: 50.0},
			wantErr: nil,
		},
		{
			name:    "negative → error",
			args:    args{amount: -50.0},
			wantErr: ErrNonPositiveAmount,
		},
		{
			name:    "zero → error",
			args:    args{amount: 0.0},
			wantErr: ErrNonPositiveAmount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateWithdrawalAmount(tt.args.amount); !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateWithdrawalAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

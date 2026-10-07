package model

import (
	"errors"
	"testing"
)

const (
	testName      = "GopherName"
	testPassword  = "GopherPassword"
	test73Symbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_"
	test72Symbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
)

func TestUser_Validate(t *testing.T) {
	type fields struct {
		Username string
		Password string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name:    "missing username → error",
			fields:  fields{Username: "", Password: testPassword},
			wantErr: ErrMissingUsername,
		},
		{
			name:    "missing password → error",
			fields:  fields{Username: testName, Password: ""},
			wantErr: ErrMissingPassword,
		},
		{
			name:    "username > 72 bytes → error",
			fields:  fields{Username: test73Symbols, Password: testPassword},
			wantErr: ErrUsernameTooLong,
		},
		{
			name:    "password > 72 bytes → error",
			fields:  fields{Username: testName, Password: test73Symbols},
			wantErr: ErrPasswordTooLong,
		},
		{
			name:    "username = 72 bytes → ok",
			fields:  fields{Username: test72Symbols, Password: testPassword},
			wantErr: nil,
		},
		{
			name:    "password = 72 bytes → ok",
			fields:  fields{Username: testName, Password: test72Symbols},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{
				Username: tt.fields.Username,
				Password: tt.fields.Password,
			}
			if err := user.Validate(); !errors.Is(err, tt.wantErr) {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

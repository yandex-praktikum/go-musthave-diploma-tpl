package config_test

import (
	"testing"

	"github.com/KalessinD/gophermart/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestValidateAddr(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
		errMsg  string // Подстрока, которая должна быть в тексте ошибки
	}{
		// Успешные кейсы
		{
			name:    "Valid address with empty host",
			addr:    ":8080",
			wantErr: false,
		},
		{
			name:    "Valid address with localhost",
			addr:    "localhost:80",
			wantErr: false,
		},
		{
			name:    "Valid address with IP",
			addr:    "127.0.0.1:443",
			wantErr: false,
		},
		{
			name:    "Valid IPv6 address",
			addr:    "[::1]:8080",
			wantErr: false,
		},

		// Ошибки
		{
			name:    "Empty address",
			addr:    "",
			wantErr: true,
			errMsg:  "address cannot be empty",
		},
		{
			name:    "Missing port",
			addr:    "localhost",
			wantErr: true,
			errMsg:  "address must be in format host:port",
		},
		{
			name:    "Invalid format (too many colons)",
			addr:    "host:80:80",
			wantErr: true,
			errMsg:  "address must be in format host:port",
		},
		{
			name:    "Non-numeric port",
			addr:    ":http",
			wantErr: true,
			errMsg:  "port must be numeric",
		},
		{
			name:    "Port too low (0)",
			addr:    ":0",
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name:    "Port too high",
			addr:    ":65536",
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name:    "Negative port",
			addr:    ":-1",
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateAddr(tt.addr)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

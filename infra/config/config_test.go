package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *AppConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid address with host and port",
			config: &AppConfig{
				RunAddress: "localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "valid IP address",
			config: &AppConfig{
				RunAddress: "192.168.1.1:8080",
			},
			wantErr: false,
		},
		{
			name: "empty address",
			config: &AppConfig{
				RunAddress: "",
			},
			wantErr:     true,
			errContains: "error parsing server address",
		},
		{
			name: "missing port",
			config: &AppConfig{
				RunAddress: "localhost",
			},
			wantErr:     true,
			errContains: "error parsing server address",
		},
		{
			name: "missing host",
			config: &AppConfig{
				RunAddress: ":",
			},
			wantErr:     true,
			errContains: "missing host",
		},
		{
			name: "only colon",
			config: &AppConfig{
				RunAddress: ":",
			},
			wantErr:     true,
			errContains: "missing host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, "localhost:8080", cfg.RunAddress)

	// Default config should be valid
	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestAppConfig_Fields(t *testing.T) {
	cfg := &AppConfig{
		RunAddress:           "localhost:8080",
		DatabaseURI:          "postgres://user:pass@localhost:5432/db",
		AccrualSystemAddress: "http://localhost:8081",
		SecretKey:            "my-secret-key",
	}

	assert.Equal(t, "localhost:8080", cfg.RunAddress)
	assert.Equal(t, "postgres://user:pass@localhost:5432/db", cfg.DatabaseURI)
	assert.Equal(t, "http://localhost:8081", cfg.AccrualSystemAddress)
	assert.Equal(t, "my-secret-key", cfg.SecretKey)

	err := cfg.Validate()
	assert.NoError(t, err)
}

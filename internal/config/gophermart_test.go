package config_test

import (
	"flag"
	"testing"

	"github.com/KalessinD/gophermart/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultGophermartConfig(t *testing.T) {
	cfg := config.GetDefaultGophermartConfig()

	assert.Equal(t, config.DefaultListenAddr, cfg.ListenAddr)
	assert.Equal(t, config.DefaultProcessingTimeout, cfg.ProcessingTimeout)
	assert.Equal(t, config.DefaultPsqlDSN, cfg.PsqlDSN)
	assert.Equal(t, config.DefaultServerEncryptionKey, cfg.EncryptionKey)
	assert.Equal(t, config.DefaultAccrualAddress, cfg.AccrualAddress)
}

func TestGophermartConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(c *config.GophermartConfig)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid default config",
			modify:  func(_ *config.GophermartConfig) {},
			wantErr: false,
		},
		{
			name: "empty encryption key",
			modify: func(c *config.GophermartConfig) {
				c.EncryptionKey = ""
			},
			wantErr: true,
			errMsg:  "encryption key can't be an empty string",
		},
		{
			name: "invalid listen address",
			modify: func(c *config.GophermartConfig) {
				c.ListenAddr = "invalid-address" // без двоеточия, обычно валидатор ругается
			},
			wantErr: true, // Предполагаем, что ValidateAddr возвращает ошибку
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.GetDefaultGophermartConfig()
			tt.modify(cfg)

			err := cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGophermartConfig_UpdateFromEnvironment(t *testing.T) {
	cfg := config.GetDefaultGophermartConfig()

	t.Setenv("RUN_ADDRESS", ":9999")
	t.Setenv("DATABASE_URI", "postgres://user:pass@localhost:6432/db")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual:8080")
	t.Setenv("GOPHERMART_ENCRYPTION_KEY", "new-key")

	err := cfg.UpdateFromEnvironment()
	require.NoError(t, err)

	assert.Equal(t, ":9999", cfg.ListenAddr)
	assert.Equal(t, "postgres://user:pass@localhost:6432/db", cfg.PsqlDSN)
	assert.Equal(t, "http://accrual:8080", cfg.AccrualAddress)
	assert.Equal(t, "new-key", cfg.EncryptionKey)
}

func TestGophermartConfig_UpdateFromCLIArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedAddr string
		expectedDSN  string
		expectedAccr string
	}{
		{
			name:         "default values if no args",
			args:         []string{},
			expectedAddr: config.DefaultListenAddr,
			expectedDSN:  config.DefaultPsqlDSN,
		},
		{
			name:         "parse custom flags",
			args:         []string{"-a", ":8081", "-d", "dsn-string", "-r", "http://acc"},
			expectedAddr: ":8081",
			expectedDSN:  "dsn-string",
			expectedAccr: "http://acc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.GetDefaultGophermartConfig()
			// Для каждого теста нужен новый FlagSet, так как флаги нельзя переопределять
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

			err := cfg.UpdateFromCLIArgs(flagSet, tt.args)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedAddr, cfg.ListenAddr)
			assert.Equal(t, tt.expectedDSN, cfg.PsqlDSN)
			assert.Equal(t, tt.expectedAccr, cfg.AccrualAddress)
		})
	}
}

func TestNewGophermartConfig(t *testing.T) {
	// Сценарий: Priority ENV -> CLI -> Default

	// Устанавливаем ENV
	t.Setenv("RUN_ADDRESS", ":7070")                   // Не должно быть перезаписано CLI
	t.Setenv("DATABASE_URI", "env-dsn")                // Должно остаться
	t.Setenv("GOPHERMART_ENCRYPTION_KEY", "valid-key") // Нужно для прохождения Validate

	// Аргументы CLI
	args := []string{"-a", ":8085"}

	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg, err := config.NewGophermartConfig(flagSet, args)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Адрес должен быть из ENV
	assert.Equal(t, ":7070", cfg.ListenAddr)
	// DSN должен быть из ENV
	assert.Equal(t, "env-dsn", cfg.PsqlDSN)
	// Ключ из ENV
	assert.Equal(t, "valid-key", cfg.EncryptionKey)
}

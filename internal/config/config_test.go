package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrderProcessInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval string
		expected time.Duration
		hasError bool
	}{
		{
			name:     "Valid 5s interval",
			interval: "5s",
			expected: 5 * time.Second,
			hasError: false,
		},
		{
			name:     "Valid 1m interval",
			interval: "1m",
			expected: 1 * time.Minute,
			hasError: false,
		},
		{
			name:     "Valid 100ms interval",
			interval: "100ms",
			expected: 100 * time.Millisecond,
			hasError: false,
		},
		{
			name:     "Invalid interval",
			interval: "invalid",
			expected: 0,
			hasError: true,
		},
		{
			name:     "Empty interval",
			interval: "",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{OrderProcessInterval: tt.interval}
			result, err := cfg.GetOrderProcessInterval()

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestLoadWithEnvironment(t *testing.T) {
	// Сохраняем оригинальные переменные окружения
	originalEnv := make(map[string]string)
	for _, key := range []string{"RUN_ADDRESS", "DATABASE_URI", "ACCRUAL_SYSTEM_ADDRESS", "ORDER_PROCESS_INTERVAL"} {
		originalEnv[key] = os.Getenv(key)
	}

	// Восстанавливаем переменные окружения после теста
	defer func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	t.Run("Load with environment variables", func(t *testing.T) {
		os.Setenv("RUN_ADDRESS", ":9090")
		os.Setenv("DATABASE_URI", "postgres://test:test@localhost/test")
		os.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual:8080")
		os.Setenv("ORDER_PROCESS_INTERVAL", "10s")

		cfg, err := loadFromValues("localhost:8080", "postgres://test:test@localhost/test", "", "5s", 5)
		require.NoError(t, err)
		assert.Equal(t, ":9090", cfg.RunAddress)
		assert.Equal(t, "postgres://test:test@localhost/test", cfg.DatabaseURI)
		assert.Equal(t, "http://accrual:8080", cfg.AccrualSystemAddress)
		assert.Equal(t, "10s", cfg.OrderProcessInterval)
	})

	t.Run("Load without DATABASE_URI", func(t *testing.T) {
		os.Unsetenv("DATABASE_URI")
		os.Unsetenv("RUN_ADDRESS")
		os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")
		os.Unsetenv("ORDER_PROCESS_INTERVAL")

		cfg, err := loadFromValues("localhost:8080", "", "", "5s", 5)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, "", cfg.DatabaseURI)
	})

	t.Run("Load with default values", func(t *testing.T) {
		os.Unsetenv("RUN_ADDRESS")
		os.Unsetenv("DATABASE_URI")
		os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")
		os.Unsetenv("ORDER_PROCESS_INTERVAL")

		cfg, err := loadFromValues("localhost:8080", "postgres://test:test@localhost/test", "", "5s", 5)
		require.NoError(t, err)
		assert.Equal(t, "localhost:8080", cfg.RunAddress)
		assert.Equal(t, "postgres://test:test@localhost/test", cfg.DatabaseURI)
		assert.Equal(t, "", cfg.AccrualSystemAddress)
		assert.Equal(t, "5s", cfg.OrderProcessInterval)
	})
}

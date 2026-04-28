package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/KalessinD/gophermart/internal/config"
)

var dotEnvLoaded = false

func TestIsProduction(t *testing.T) {
	dotEnvLoaded = true

	t.Run("Is Production", func(t *testing.T) {
		t.Setenv("YP_ENV", "prod")
		defer os.Unsetenv("YP_ENV")
		assert.True(t, config.IsProduction())
		assert.False(t, config.IsDevelopment())
	})

	t.Run("Is Development (env set to dev)", func(t *testing.T) {
		t.Setenv("YP_ENV", "dev")
		defer os.Unsetenv("YP_ENV")
		assert.False(t, config.IsProduction())
		assert.True(t, config.IsDevelopment())
	})

	t.Run("Is Development (env not set)", func(t *testing.T) {
		os.Unsetenv("YP_ENV")
		assert.False(t, config.IsProduction())
		assert.True(t, config.IsDevelopment())
	})
}

func TestGetEnvOrFallback(t *testing.T) {
	t.Run("String type", func(t *testing.T) {
		t.Setenv("STR_VAR", "hello")
		defer os.Unsetenv("STR_VAR")
		val := config.GetEnvOrFallback("STR_VAR", "default")
		assert.Equal(t, "hello", val)
	})

	t.Run("String fallback", func(t *testing.T) {
		os.Unsetenv("STR_VAR_EMPTY")
		val := config.GetEnvOrFallback("STR_VAR_EMPTY", "default")
		assert.Equal(t, "default", val)
	})

	t.Run("Int type", func(t *testing.T) {
		t.Setenv("INT_VAR", "123")
		defer os.Unsetenv("INT_VAR")
		val := config.GetEnvOrFallback("INT_VAR", 0)
		assert.Equal(t, 123, val)
	})

	t.Run("Int invalid format", func(t *testing.T) {
		t.Setenv("INT_BAD", "not-an-int")
		defer os.Unsetenv("INT_BAD")
		val := config.GetEnvOrFallback("INT_BAD", 99)
		assert.Equal(t, 99, val, "Should return fallback on parse error")
	})

	t.Run("Float type", func(t *testing.T) {
		t.Setenv("FLOAT_VAR", "3.14")
		defer os.Unsetenv("FLOAT_VAR")
		val := config.GetEnvOrFallback("FLOAT_VAR", 0.0)
		assert.Equal(t, 3.14, val)
	})
}

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Init(t *testing.T) {
	t.Run("full_initialization_success", func(t *testing.T) {
		args := []string{"-d", "postgres://dsn", "-a", "localhost:1111", "-r", "localhost:2222"}

		cfg, err := NewConfig().
			SetJWTSecretKey("secret").
			Init(
				WithEnv(), // ENV пустые
				WithCmdArgs(&args),
			)

		require.NoError(t, err)
		assert.Equal(t, "postgres://dsn", cfg.DBDSN())
		assert.Equal(t, "localhost:1111", cfg.ListenAddr())
		assert.Equal(t, "localhost:2222", cfg.AccrualSystemAddr())
	})

	t.Run("priority_test_flags_over_env", func(t *testing.T) {
		args := []string{"-d", "flag-dsn"}

		// Создаем базу и подменяем поиск ENV
		cfgBase := NewConfig()
		cfgBase.lookupEnv = func(key string) (string, bool) {
			if key == "DATABASE_URI" {
				return "env-dsn", true
			}
			return "", false
		}

		cfg, err := cfgBase.
			SetJWTSecretKey("secret").
			Init(
				WithEnv(),          // Сначала ставит env-dsn
				WithCmdArgs(&args), // Флаг перекрывает на flag-dsn
			)

		require.NoError(t, err)
		assert.Equal(t, "flag-dsn", cfg.DBDSN())
	})

	t.Run("validation_error_missing_secret", func(t *testing.T) {
		args := []string{"-d", "dsn"}
		_, err := NewConfig().Init(WithCmdArgs(&args))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT secret key is required")
	})

	t.Run("immutable_map_protection", func(t *testing.T) {
		cfg := NewConfig()
		m1 := cfg.CompressibleContentTypes()
		m1["application/xml"] = struct{}{} // Пытаемся сломать извне

		m2 := cfg.CompressibleContentTypes()
		assert.NotContains(t, m2, "application/xml", "Internal map should remain protected")
	})
}

func TestSocketAddr(t *testing.T) {
	t.Run("valid_addr", func(t *testing.T) {
		addr, err := NewSocketAddr("127.0.0.1:8080")
		require.NoError(t, err)
		assert.Equal(t, "127.0.0.1", addr.hostname)
		assert.Equal(t, "8080", addr.port)
	})

	t.Run("invalid_addr", func(t *testing.T) {
		_, err := NewSocketAddr("localhost") // Ошибка: нет порта
		assert.Error(t, err)
	})
}

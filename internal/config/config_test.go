package config_test

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-musthave-diploma-tpl/internal/config"
)

func TestLoad_FromEnv(t *testing.T) {
	// Устанавливаем переменные окружения
	os.Setenv("RUN_ADDRESS", ":9090")
	os.Setenv("DATABASE_URI", "postgresql://user:pass@localhost:5432/testdb?sslmode=disable")
	os.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual:8081")
	defer func() {
		os.Unsetenv("RUN_ADDRESS")
		os.Unsetenv("DATABASE_URI")
		os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")
	}()

	// Сбрасываем флаги для чистоты теста
	os.Args = []string{"test"}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	cfg, err := config.LoadWithFlagSet(fs)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, ":9090", cfg.RunAddress)
	assert.Equal(t, "postgresql://user:pass@localhost:5432/testdb?sslmode=disable", cfg.DatabaseURI)
	assert.Equal(t, "http://accrual:8081", cfg.AccrualSystemAddress)
}

func TestLoad_FlagsPriority(t *testing.T) {
	// Устанавливаем переменные окружения
	os.Setenv("RUN_ADDRESS", ":9090")
	os.Setenv("DATABASE_URI", "postgresql://env:pass@localhost:5432/envdb?sslmode=disable")
	os.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://env:8081")
	defer func() {
		os.Unsetenv("RUN_ADDRESS")
		os.Unsetenv("DATABASE_URI")
		os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")
	}()

	// Создаём новый набор флагов для теста (избегаем переопределения)
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = []string{
		"test",
		"-a", ":3000",
		"-d", "postgresql://flag:pass@localhost:5432/flagdb?sslmode=disable",
		"-r", "http://flag:8082",
	}

	cfg, err := config.LoadWithFlagSet(fs)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Флаги должны переопределить переменные окружения
	assert.Equal(t, ":3000", cfg.RunAddress)
	assert.Equal(t, "postgresql://flag:pass@localhost:5432/flagdb?sslmode=disable", cfg.DatabaseURI)
	assert.Equal(t, "http://flag:8082", cfg.AccrualSystemAddress)
}

func TestLoad_Defaults(t *testing.T) {
	// Очищаем переменные окружения
	os.Unsetenv("RUN_ADDRESS")
	os.Unsetenv("DATABASE_URI")
	os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")
	os.Args = []string{"test"}

	// Создаём новый FlagSet для избежания переопределения флагов
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	// DATABASE_URI не установлен — должна быть ошибка
	cfg, err := config.LoadWithFlagSet(fs)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "DATABASE_URI или флаг -d обязателен")
}

func TestLoad_WithRequiredDatabaseURI(t *testing.T) {
	// Устанавливаем только обязательный DATABASE_URI
	os.Setenv("DATABASE_URI", "postgresql://user:pass@localhost:5432/testdb?sslmode=disable")
	defer os.Unsetenv("DATABASE_URI")
	os.Args = []string{"test"}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	cfg, err := config.LoadWithFlagSet(fs)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Проверяем значения по умолчанию
	assert.Equal(t, ":8080", cfg.RunAddress)
	assert.Equal(t, "postgresql://user:pass@localhost:5432/testdb?sslmode=disable", cfg.DatabaseURI)
	assert.Equal(t, "http://localhost:8080", cfg.AccrualSystemAddress)
}

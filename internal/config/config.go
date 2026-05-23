package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

// Config содержит конфигурацию приложения.
type Config struct {
	// RunAddress — адрес и порт запуска HTTP-сервера.
	RunAddress string
	// DatabaseURI — URI подключения к PostgreSQL.
	DatabaseURI string
	// AccrualSystemAddress — адрес системы расчёта начислений.
	AccrualSystemAddress string
}

// Load загружает конфигурацию из переменных окружения и флагов командной строки.
// Флаги командной строки имеют приоритет над переменными окружения.
func Load() (*Config, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	return LoadWithFlagSet(fs)

}

// LoadWithFlagSet загружает конфигурацию с использованием указанного набора флагов.
// Эта функция полезна для тестирования.
func LoadWithFlagSet(fs *flag.FlagSet) (*Config, error) {
	cfg := &Config{}

	// Определяем флаги
	fs.StringVar(&cfg.RunAddress, "a", getEnvDefault("RUN_ADDRESS", ":8080"), "Адрес и порт запуска HTTP-сервера")
	fs.StringVar(&cfg.DatabaseURI, "d", getEnvDefault("DATABASE_URI", ""), "URI подключения к PostgreSQL")
	fs.StringVar(&cfg.AccrualSystemAddress, "r", getEnvDefault("ACCRUAL_SYSTEM_ADDRESS", "http://localhost:8080"), "Адрес системы расчёта начислений")

	// Парсим флаги
	err := fs.Parse(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга флагов: %w", err)
	}

	// Валидация обязательных параметров
	if cfg.DatabaseURI == "" {
		return nil, errors.New("DATABASE_URI или флаг -d обязателен")
	}

	return cfg, nil
}

// getEnvDefault возвращает значение переменной окружения или значение по умолчанию.
func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

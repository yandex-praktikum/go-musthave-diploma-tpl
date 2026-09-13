// Package config отвечает за загрузку и разбор конфигурации сервиса.
//
// Конфигурация собирается из переменных окружения и флагов командной строки,
// при этом флаги имеют приоритет над переменными окружения.
package config

import (
	"errors"
	"flag"
	"os"
	"runtime"
	"strings"
)

// Config содержит все параметры конфигурации сервиса Gophermart.
type Config struct {
	// RunAddress — адрес и порт, на которых запускается HTTP-сервер.
	RunAddress string
	// DatabaseURI — строка подключения к базе данных PostgreSQL.
	DatabaseURI string
	// AccrualSystemAddress — адрес внешней системы расчёта баллов (accrual).
	AccrualSystemAddress string
	// SecretKey — секретный ключ для подписи и проверки JWT-токенов.
	SecretKey string
}

// ErrSecretKeyRequired возвращается, если переменная окружения SECRET_KEY
// не задана. Секретный ключ обязателен: без него JWT-токены будут
// подписываться предсказуемым ключом, что позволит злоумышленнику
// подделывать токены.
var ErrSecretKeyRequired = errors.New("SECRET_KEY environment variable is required")

// Load загружает конфигурацию сервиса.
//
// Значения по умолчанию берутся из переменных окружения RUN_ADDRESS,
// DATABASE_URI и ACCRUAL_SYSTEM_ADDRESS, а затем могут быть переопределены
// флагами командной строки -a, -d и -r соответственно.
//
// Секретный ключ для JWT берётся из переменной SECRET_KEY и является
// обязательным: если переменная не задана, возвращается ErrSecretKeyRequired.
//
// На Windows в строке подключения к БД "localhost" заменяется на "127.0.0.1"
// (кроме случая с host.docker.internal), так как localhost там может
// резолвиться в IPv6 (::1), а PostgreSQL часто слушает только IPv4.
//
// Возвращает указатель на заполненную конфигурацию или ошибку.
func Load() (*Config, error) {
	cfg := &Config{}

	// 1. Сначала читаем переменные окружения как значения по умолчанию
	cfg.RunAddress = getEnv("RUN_ADDRESS", "localhost:8080")
	cfg.DatabaseURI = getEnv("DATABASE_URI", "")
	cfg.AccrualSystemAddress = getEnv("ACCRUAL_SYSTEM_ADDRESS", "")

	// 2. Регистрируем флаги со значениями по умолчанию из env
	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "address to run server")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "database URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "accrual system address")
	flag.Parse()

	// 3. Секретный ключ обязателен — fail early, если не задан
	cfg.SecretKey = os.Getenv("SECRET_KEY")
	if cfg.SecretKey == "" {
		return nil, ErrSecretKeyRequired
	}

	// 4. Нормализация адреса БД для Windows
	if runtime.GOOS == "windows" && !strings.Contains(cfg.DatabaseURI, "host.docker.internal") {
		cfg.DatabaseURI = strings.Replace(cfg.DatabaseURI, "localhost", "127.0.0.1", -1)
	}

	return cfg, nil
}

// getEnv возвращает значение переменной окружения или значение по умолчанию,
// если переменная не задана или пуста.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

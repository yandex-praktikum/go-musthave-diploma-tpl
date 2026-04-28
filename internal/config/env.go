package config

import (
	"os"
	"strconv"
)

// Returns true if YP_ENV environment is set to "prod", which means the production environment.
//
// Otherwise returns false.
func IsProduction() bool {
	return GetEnv("YP_ENV") == "prod"
}

// Returns true if YP_ENV environment is not set to "prod", which means the production environment.
//
// Otherwise returns false.
func IsDevelopment() bool {
	return !IsProduction()
}

// Returns the value of os.GetEnv if exists
func GetEnv(key string) string {
	return os.Getenv(key)
}

// Returns OS Environment
func GetEnvOrFallback[T any](key string, fallback T) T {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}

	// Нам нужно привести строку к T.
	// Используем type assertion через any, чтобы проверить возможные типы.
	var result any
	var err error

	// Пытаемся определить тип T и распарсить строку
	switch any(fallback).(type) {
	case string:
		result = valStr
	case int:
		result, err = strconv.Atoi(valStr)
	case float64:
		result, err = strconv.ParseFloat(valStr, 64)
	default:
		// Если тип не поддерживается, возвращаем fallback
		return fallback
	}

	if err != nil {
		return fallback // Если ошибка парсинга, возвращаем дефолт
	}

	if val, ok := result.(T); ok {
		return val
	}

	return fallback
}

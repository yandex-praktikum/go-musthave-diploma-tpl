package config

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/IvanDolgov/go-musthave-diploma-tpl/internal/models"
)

// ParseServerFlags парсит флаги командной строки для сервера
func ParseServerFlags() models.Config {
	var (
		address     string
		databaseDsn string
		key         string // секретный ключ
	)

	// Регистрируем флаги для сервера
	flag.StringVar(&address, "a", "localhost:8080", "server address")
	flag.StringVar(&databaseDsn, "d", "", "database_dsn")
	flag.StringVar(&key, "k", "", "secret key for request signing")

	flag.Parse()

	// Применяем приоритеты параметров (env vars имеют приоритет над флагами)
	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		address = envAddress
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		databaseDsn = envDatabaseDSN
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		key = envKey
	}

	// Парсим адрес на server и port
	server, port := parseAddress(address)

	return models.Config{
		Address:     address,
		Server:      server,
		Port:        port,
		DatabaseDSN: databaseDsn,
		Key:         key,
	}
}

func getEnvBool(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultVal
}

func getEnvInt64(key string, defaultVal int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// parseAddress разбивает адрес на сервер и порт
func parseAddress(address string) (string, string) {
	parts := strings.Split(address, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], "8080" // порт по умолчанию
}

package config

import (
	"flag"
	"github.com/sirupsen/logrus"
	"os"
)

type Config struct {
	ServerAddress  string
	AccrualAddress string
	DBdsn          string
	MigrationsPath string
	LogLevel       logrus.Level
	SecretKey      string
}

func LoadConfig() *Config {
	cfg := Config{
		ServerAddress: "localhost:8081",
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
	flag.StringVar(&cfg.AccrualAddress, "r", cfg.AccrualAddress, "Accrual system address")
	flag.StringVar(&cfg.DBdsn, "d", cfg.DBdsn, "Postgres DSN")
	flag.StringVar(&cfg.SecretKey, "k", cfg.SecretKey, "Secret sign key")
	flag.Parse()

	loadEnvVariables(&cfg)

	return &cfg
}

func loadEnvVariables(cfg *Config) {
	if addr := os.Getenv("RUN_ADDRESS"); addr != "" {
		cfg.ServerAddress = addr
	}
	if accrualAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accrualAddress != "" {
		cfg.AccrualAddress = accrualAddress
	}
	if dbDSN := os.Getenv("DATABASE_URI"); dbDSN != "" {
		cfg.DBdsn = dbDSN
	}
	if secretKey := os.Getenv("KEY"); secretKey != "" {
		cfg.SecretKey = secretKey
	}
	loggingLevel := getEnv("LOGGING_LEVEL", "debug")
	level, err := logrus.ParseLevel(loggingLevel)
	if err != nil {
		logrus.Fatalf("Invalid log level: %s", loggingLevel)
	}
	cfg.LogLevel = level
	migrationsPath := getEnv("MIGRATION_PATH", "internal/storage/postgres/migrate")
	cfg.MigrationsPath = migrationsPath
}

func getEnv(key string, defaultValue ...string) string {
	value := os.Getenv(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

func New() *Config {
	var (
		runAddr = flag.String("a", "", "адрес и порт запуска сервиса")
		dbURI   = flag.String("d", "", "адрес подключения к базе данных")
		accrual = flag.String("r", "", "адрес системы расчёта начислений")
	)
	flag.Parse()

	cfg := &Config{
		RunAddress:           getEnvOrDefault("RUN_ADDRESS", *runAddr),
		DatabaseURI:          getEnvOrDefault("DATABASE_URI", *dbURI),
		AccrualSystemAddress: getEnvOrDefault("ACCRUAL_SYSTEM_ADDRESS", *accrual),
	}
	return cfg
}

func getEnvOrDefault(env, fallback string) string {
	if val := os.Getenv(env); val != "" {
		return val
	}
	return fallback
}

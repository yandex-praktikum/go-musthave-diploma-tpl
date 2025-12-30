package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress        string
	DatabaseURI       string
	AccrualSystemAddr string
}

func Load() *Config {
	var cfg Config

	flag.StringVar(&cfg.RunAddress, "a", getEnvDefault("RUN_ADDRESS", "localhost:8080"), "HTTP listen address")
	flag.StringVar(&cfg.DatabaseURI, "d", getEnvDefault("DATABASE_URI", ""), "PostgreSQL DSN")
	flag.StringVar(&cfg.AccrualSystemAddr, "r", getEnvDefault("ACCRUAL_SYSTEM_ADDRESS", ""), "accrual system base URL")

	flag.Parse()

	return &cfg
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

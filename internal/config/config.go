package config

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	RunAddress        string
	DatabaseURI       string
	AccrualSystemAddr string
	JWTSecret         string
}

func Load() *Config {
	var cfg Config

	flag.StringVar(&cfg.RunAddress, "a", getEnvDefault("RUN_ADDRESS", "localhost:8080"), "HTTP listen address")
	flag.StringVar(&cfg.DatabaseURI, "d", getEnvDefault("DATABASE_URI", ""), "PostgreSQL DSN")
	flag.StringVar(&cfg.AccrualSystemAddr, "r", getEnvDefault("ACCRUAL_SYSTEM_ADDRESS", ""), "accrual system base URL")

	flag.StringVar(&cfg.JWTSecret, "jwt-secret", getEnvDefault("JWT_SECRET", "dev-secret-key-change-in-production"), "JWT secret key")

	flag.Parse()

	if cfg.JWTSecret == "dev-secret-key-change-in-production" {
		log.Println("WARNING: Using default JWT secret for development. Set JWT_SECRET env variable for production!")
	}

	return &cfg
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

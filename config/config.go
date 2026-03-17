package config

import (
	"flag"
	"os"
	"strings"
)

const (
	RUN_ADDRESS            = "RUN_ADDRESS"
	DATABASE_URI           = "DATABASE_URI"
	ACCRUAL_SYSTEM_ADDRESS = "ACCRUAL_SYSTEM_ADDRESS"
)

type AppConfig struct {
	Host           string `env:"SERVER_ADDRESS"`
	DataBaseString string `env:"BASE_URL"`
	AccrualSys     string `env:"FILE_STORAGE_PATH"`
}

var cfg AppConfig

func NewCfg() *AppConfig {

	cfq := &AppConfig{}

	flag.StringVar(&cfq.Host, "a", "localhost:8080", "Host")
	flag.StringVar(&cfq.DataBaseString, "d", "DB", "Result URL")
	flag.StringVar(&cfq.AccrualSys, "r", "AccrualSys", "FilePATH")

	flag.Parse()

	if host := os.Getenv(RUN_ADDRESS); host != "" {
		cfq.Host = strings.TrimSpace(host)
	}
	if dburl := os.Getenv(DATABASE_URI); dburl != "" {
		cfq.DataBaseString = strings.TrimSpace(dburl)
	}
	if accrualsys := os.Getenv(ACCRUAL_SYSTEM_ADDRESS); accrualsys != "" {
		cfq.AccrualSys = strings.TrimSpace(accrualsys)
	}

	return cfq
}

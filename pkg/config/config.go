package config

import (
	"crypto/rand"
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddr         string `env:"RUN_ADDRESS"`
	DataBaseURI     string `env:"DATABASE_URI"`
	Accural         string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	CookieSecretKey []byte `env:"COOKIE_SECRET_KEY"`
}

var (
	flagRunAddr     = flag.String("a", "", "address and port to run server")
	flagDataBaseURI = flag.String("d", "", "DSN to connect to the database")
	flagAccural     = flag.String("r", "", "address of the system calculation calculations")
)

func NewConfig() (*Config, error) {
	flag.Parse()
	cfg := &Config{}

	// Парсим переменные окружения
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env: %w", err)
	}

	// Если флаг передан, перезаписываем значения
	if *flagRunAddr != "" {
		cfg.RunAddr = *flagRunAddr
	}

	if *flagDataBaseURI != "" {
		cfg.DataBaseURI = *flagDataBaseURI
	}

	if *flagAccural != "" {
		cfg.Accural = *flagAccural
	}

	// Устанавливаем значение по умолчанию
	if cfg.RunAddr == "" {
		cfg.RunAddr = ":8080"
	} else if !strings.Contains(cfg.RunAddr, ":") {
		cfg.RunAddr = ":" + cfg.RunAddr
	}

	// Генерируем ключ ТОЛЬКО если он не задан через ENV
	if len(cfg.CookieSecretKey) == 0 {
		cfg.CookieSecretKey = GenerateKeyToken()
	}
	return cfg, nil
}

func GenerateKeyToken() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	return key
}

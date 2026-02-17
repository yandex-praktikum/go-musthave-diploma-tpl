// Package config содержит конфигурацию приложения
package config

import (
	"flag"

	env "github.com/caarlos0/env/v11"
)

type Config struct {
	SelfAddress          string `env:"RUN_ADDRESS"`
	DSN                  string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	ENV                  string `env:"ENV"`

	PrivateKeyJWT string `env:"PRIVATE_KEY_JWT"`
	PrivateKey    string `env:"PRIVATE_KEY"`
}

func New() (*Config, error) {
	cfg := &Config{}

	// Парсинг флагов командной строки
	flag.StringVar(&cfg.SelfAddress, "a", ":8080", "Адрес и порт запуска сервиса")
	flag.StringVar(&cfg.DSN, "d", "", "Адрес подключения к базе данных")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Адрес системы расчёта начислений")
	flag.Parse()

	// Переменные окружения с более высоким приоритетом
	envCfg := &Config{}
	if err := env.Parse(envCfg); err != nil {
		return nil, err
	}

	// Перезаписываем значения из переменных окружения, если они заданы
	if envCfg.SelfAddress != "" {
		cfg.SelfAddress = envCfg.SelfAddress
	}
	if envCfg.DSN != "" {
		cfg.DSN = envCfg.DSN
	}
	if envCfg.AccrualSystemAddress != "" {
		cfg.AccrualSystemAddress = envCfg.AccrualSystemAddress
	}
	if envCfg.ENV != "" {
		cfg.ENV = envCfg.ENV
	}
	if envCfg.PrivateKeyJWT != "" {
		cfg.PrivateKeyJWT = envCfg.PrivateKeyJWT
	}
	if envCfg.PrivateKey != "" {
		cfg.PrivateKey = envCfg.PrivateKey
	}

	return cfg, nil
}

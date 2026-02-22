// Package config содержит конфигурацию приложения
// Приоритет отдан переменным окружения. Если они не заданы, конфигурация задается через флаги
// или устанавливается дефолтное значение
package config

import (
	"flag"
	"sync"

	env "github.com/caarlos0/env/v11"
)

// можно обойтись без этого, но в тестировании были проблемы
var (
	flagOnce    sync.Once
	flagAddress string
	flagDSN     string
	flagAccrual string
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

	// Парсинг флагов командной строки (регистрируем только один раз)
	flagOnce.Do(func() {
		flag.StringVar(&flagAddress, "a", ":8080", "Адрес и порт запуска сервиса")
		flag.StringVar(&flagDSN, "d", "", "Адрес подключения к базе данных")
		flag.StringVar(&flagAccrual, "r", "", "Адрес системы расчёта начислений")
		flag.Parse()
	})

	cfg.SelfAddress = flagAddress
	cfg.DSN = flagDSN
	cfg.AccrualSystemAddress = flagAccrual

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

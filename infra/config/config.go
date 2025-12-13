package config

import (
	"fmt"
	"net"
)

type AppConfig struct {
	RunAddress           string `yaml:"run_address"`            //Адрес и порт запуска сервиса
	DatabaseURI          string `yaml:"database_uri"`           //Адрес подключения к базе данных
	AccrualSystemAddress string `yaml:"accrual_system_address"` //Адрес системы расчёта начислений
	SecretKey            string `yaml:"secret_key"`             //Секретный ключ для подписи сессионного токена
}

func NewDefaultConfig() *AppConfig {

	return &AppConfig{

		RunAddress: "localhost:8080",
	}
}

func (c *AppConfig) Validate() error {
	// Checks the format without resolving the hostname
	host, port, err := net.SplitHostPort(c.RunAddress)
	if err != nil {
		return fmt.Errorf("error parsing server address: %w", err)
	}
	if host == "" {
		return fmt.Errorf("missing host in address %q", c.RunAddress)
	}
	if port == "" {
		return fmt.Errorf("missing port in address %q", c.RunAddress)
	}
	return nil
}

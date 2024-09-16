package main

import (
	"flag"
	"os"
)

type Configs struct {
	RunAddress           string
	LogLevel             string
	AddrConDB            string
	AccrualSystemAddress string
}

func NewConfig() *Configs {
	return &Configs{}
}

func (c *Configs) Parsed() {
	c.parseFlags()
	// Проверка переменной окружения RUN_ADDRESS
	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		c.RunAddress = envRunAddress
	}

	// Проверка переменной окружения LOG_LEVEL
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		c.LogLevel = envLogLevel
	}

	// Проверка переменной окружения DATABASE_URI
	if envAddrConDB := os.Getenv("DATABASE_URI"); envAddrConDB != "" {
		c.AddrConDB = envAddrConDB
	}

	// Проверка переменной окружения ACCRUAL_SYSTEM_ADDRESS
	if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
		c.AccrualSystemAddress = envAccrualSystemAddress
	}
}

func (c *Configs) parseFlags() {
	// Флаг -a отвечает за адрес запуска HTTP-сервера (значение может быть таким: localhost:8080).
	flag.StringVar(&c.RunAddress, "a", ":8080", "Server address host:port")
	// Флаг -l отвечает за logger
	flag.StringVar(&c.LogLevel, "l", "info", "log level")
	// Флаг -p отвечает за адрес подключения DB
	flag.StringVar(&c.AddrConDB, "d", "postgresql://shortner:123456789@localhost:5432/postgres?sslmode=disable", "address DB")
	// Флаг -r отвечает за адрес системы расчета начислений
	flag.StringVar(&c.AccrualSystemAddress, "r", "", "address accrual system address")
	flag.Parse()
}

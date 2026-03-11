package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	RunAddress           string
	DatabaseDSN          string
	AccrualSystemAddress string
	AccrualPollInterval  time.Duration
	LogLevel             string
	CookieSecret         string
}

// Option — функция для настройки конфигурации.
type Option func(*Config)

// WithRunAddress устанавливает адрес сервера.
func WithRunAddress(addr string) Option {
	return func(c *Config) {
		c.RunAddress = addr
	}
}

// WithDatabaseDSN устанавливает строку подключения к БД.
func WithDatabaseDSN(dsn string) Option {
	return func(c *Config) {
		c.DatabaseDSN = dsn
	}
}

// WithAccrualSystemAddress устанавливает адрес системы начислений.
func WithAccrualSystemAddress(addr string) Option {
	return func(c *Config) {
		c.AccrualSystemAddress = addr
	}
}

// WithAccrualPollInterval устанавливает интервал опроса системы начислений.
func WithAccrualPollInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.AccrualPollInterval = interval
	}
}

// WithLogLevel устанавливает уровень логирования.
func WithLogLevel(level string) Option {
	return func(c *Config) {
		c.LogLevel = level
	}
}

// WithCookieSecret устанавливает секрет для cookies.
func WithCookieSecret(secret string) Option {
	return func(c *Config) {
		c.CookieSecret = secret
	}
}

// New создаёт конфигурацию с дефолтными значениями и применяет опции.
func New(opts ...Option) *Config {
	c := &Config{
		RunAddress:           "localhost:8080",
		LogLevel:             "info",
		DatabaseDSN:          "postgres://shortener:shortener@localhost:5432/shortener",
		CookieSecret:         "dev-secret",
		AccrualSystemAddress: "",
		AccrualPollInterval:  2 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func ParseEnv(config *Config) {
	if RunAddress := os.Getenv("SERVER_ADDRESS"); RunAddress != "" {
		config.RunAddress = RunAddress
	}
	if LogLevel := os.Getenv("LOG_LEVEL"); LogLevel != "" {
		config.LogLevel = LogLevel
	}
	if DatabaseDSN := os.Getenv("DATABASE_DSN"); DatabaseDSN != "" {
		config.DatabaseDSN = DatabaseDSN
	}
	if s := os.Getenv("COOKIE_SECRET"); s != "" {
		config.CookieSecret = s
	}
	if s := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); s != "" {
		config.AccrualSystemAddress = s
	}
}

func ParseFlags(config *Config) {
	flag.StringVar(&config.RunAddress, "a", config.RunAddress, "address and port to run server")
	flag.StringVar(&config.LogLevel, "l", config.LogLevel, "log level")
	flag.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "database connection string")
	flag.StringVar(&config.AccrualSystemAddress, "r", config.AccrualSystemAddress, "accrual system address")

	flag.Parse()
}

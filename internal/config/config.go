package config

import (
	"flag"
	"os"
	"strconv"
	"time"
)

// Config содержит конфигурацию сервера
type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	OrderProcessInterval string
	WorkerCount          int
}

// GetOrderProcessInterval возвращает интервал обработки заказов как time.Duration
func (c *Config) GetOrderProcessInterval() (time.Duration, error) {
	return time.ParseDuration(c.OrderProcessInterval)
}

// Load загружает конфигурацию из флагов и переменных окружения
func Load() (*Config, error) {
	var (
		flagRunAddress           string
		flagDatabaseURI          string
		flagAccrualSystemAddress string
		flagOrderProcessInterval string
		flagWorkerCount          int
	)

	flag.StringVar(&flagRunAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagDatabaseURI, "d", "", "database URI")
	flag.StringVar(&flagAccrualSystemAddress, "r", "", "accrual system address")
	flag.StringVar(&flagOrderProcessInterval, "i", "5s", "order processing interval")
	flag.IntVar(&flagWorkerCount, "w", 5, "number of workers for order processing")
	flag.Parse()

	return loadFromValues(flagRunAddress, flagDatabaseURI, flagAccrualSystemAddress, flagOrderProcessInterval, flagWorkerCount)
}

// loadFromValues загружает конфигурацию из переданных значений
func loadFromValues(runAddress, databaseURI, accrualSystemAddress, orderProcessInterval string, workerCount int) (*Config, error) {
	// Приоритет: flag > env > default
	if runAddress == "localhost:8080" {
		if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
			runAddress = envRunAddress
		}
	}
	if databaseURI == "" {
		if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
			databaseURI = envDatabaseURI
		}
	}
	if accrualSystemAddress == "" {
		if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
			accrualSystemAddress = envAccrualSystemAddress
		}
	}
	if orderProcessInterval == "5s" {
		if envOrderProcessInterval := os.Getenv("ORDER_PROCESS_INTERVAL"); envOrderProcessInterval != "" {
			orderProcessInterval = envOrderProcessInterval
		}
	}
	if workerCount == 5 {
		if envWorkerCount := os.Getenv("WORKER_COUNT"); envWorkerCount != "" {
			if parsed, err := strconv.Atoi(envWorkerCount); err == nil {
				workerCount = parsed
			}
		}
	}

	return &Config{
		RunAddress:           runAddress,
		DatabaseURI:          databaseURI,
		AccrualSystemAddress: accrualSystemAddress,
		OrderProcessInterval: orderProcessInterval,
		WorkerCount:          workerCount,
	}, nil
}

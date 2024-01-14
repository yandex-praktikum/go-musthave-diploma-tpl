package accrual

import (
	"github.com/iRootPro/gophermart/internal/store/sqlstore"
	"time"
)

type Config struct {
	RunAddress     string
	LogLevel       string
	PoolingTimeout time.Duration
	store          *sqlstore.Store
}

func NewAccrualConfig(address string, loglevel string) *Config {
	return &Config{
		RunAddress:     address,
		LogLevel:       loglevel,
		PoolingTimeout: 5 * time.Second,
	}
}

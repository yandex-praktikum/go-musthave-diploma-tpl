package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
)

type GophermartConfig struct {
	TokenExp              time.Duration `env:"JWT_EXP"`
	TokenSecret           string        `env:"JWT_SECRET"`
	DatabaseUri           string        `env:"DATABASE_URI"`
	RunAddress            string        `env:"RUN_ADDRESS"`
	AccrualSystemAddress  string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
	AcrualSystemPoolCount int           `env:"ACRUAL_SYSTEM_POOL_COUNT"`
	MaxConns              int           `env:"DATABASE_MAX_CONNS"`
	MaxConnLifetime       time.Duration `env:"DATABASE_MAX_CONN_LIFE_TIME"`
	MaxConnIdleTime       time.Duration `env:"DATABASE_MAX_CONN_IDLE_TIME"`
	ProcessingLimit       int           `env:"ORDER_PROCESSING_LIMIT"`
	ProcessingScoreDelta  time.Duration `env:"ORDER_PROCESSING_DELTA"`
}

func LoadGophermartConfig() (*GophermartConfig, error) {
	srvConf := &GophermartConfig{}

	flag.StringVar(&srvConf.TokenSecret, "jwtSec", "secret", "jwt secret")
	flag.DurationVar(&srvConf.TokenExp, "jwtExp", 3*time.Hour, "jwt expiration period")
	flag.IntVar(&srvConf.MaxConns, "dbMaxConns", 5, "database max connections")
	flag.IntVar(&srvConf.AcrualSystemPoolCount, "accrualPoolCount", 5, "accrual pool count")
	flag.DurationVar(&srvConf.MaxConnLifetime, "dbMaxConnLifeTime", 5*time.Minute, "database max connection life time")
	flag.DurationVar(&srvConf.MaxConnIdleTime, "dbMaxConnIdleTime", 5*time.Minute, "database max connection idle time")

	flag.IntVar(&srvConf.ProcessingLimit, "pLimit", 10, "order processing limit")
	flag.DurationVar(&srvConf.ProcessingScoreDelta, "pDelta", 20*time.Second, "order processing delta")

	flag.StringVar(&srvConf.RunAddress, "a", ":8080", "server address (format \":PORT\")")
	flag.StringVar(&srvConf.DatabaseUri, "d", "", "PostgreSQL URL like 'postgres://username:password@localhost:5432/database_name'")
	flag.StringVar(&srvConf.AccrualSystemAddress, "r", "localhost:8080", "server address (format \":PORT\")")

	flag.Parse()

	err := env.Parse(srvConf)
	if err != nil {
		return nil, err
	}

	return srvConf, nil
}

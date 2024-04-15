package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
)

type GophermartConfig struct {
	TokenExp             time.Duration `env:"JWT_EXP"`
	TokenSecret          string        `env:"JWT_SECRET"`
	DatabaseUri          string        `env:"DATABASE_URI"`
	RunAddress           string        `env:"RUN_ADDRESS"`
	AccrualSystemAddress string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
	MaxConns             int           `env:"DATABASE_MAX_CONNS"`
}

func LoadGophermartConfig() (*GophermartConfig, error) {
	srvConf := &GophermartConfig{}

	flag.StringVar(&srvConf.TokenSecret, "jwtSec", "secret", "jwt secret")
	flag.DurationVar(&srvConf.TokenExp, "jwtExp", 3*time.Hour, "jwt expiration period")
	flag.IntVar(&srvConf.MaxConns, "dbMaxCon", 5, "database max connections")

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

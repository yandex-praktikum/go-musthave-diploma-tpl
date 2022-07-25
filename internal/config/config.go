package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DataBaseDsn          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func GetConfig() (Config, error) {
	cfg := Config{}

	flag.StringVar(&cfg.RunAddress, "a", ":8080", "port start listen")
	flag.StringVar(&cfg.DataBaseDsn, "d", "postgresql://postgres:sqllife@localhost:5434/gophermart", "database dsn")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", ":8080", "AccrualSystemAddress")
	flag.Parse()
	//

	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

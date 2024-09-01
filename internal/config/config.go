package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	SecretKey            string `env:"SECRET_KEY"`
}

func NewConfig() *Config {
	config := new(Config)
	config.ReadFlags()
	config.ReadEnv()
	return config
}

func (ac *Config) ReadFlags() {
	flag.StringVar(&ac.RunAddress, "a", "localhost:8000", "app address and port")
	flag.StringVar(&ac.DatabaseURI, "d", "host=localhost user=postgres password=351762 dbname=gophermart sslmode=disable", "database uri")
	flag.StringVar(&ac.AccrualSystemAddress, "r", "http://localhost:8080", "accrual system address and port")
	flag.StringVar(&ac.SecretKey, "sk", "supersecretkey", "secret key")
	flag.Parse()
}

func (ac *Config) ReadEnv() {
	err := env.Parse(ac)
	if err != nil {
		panic(err)
	}
}

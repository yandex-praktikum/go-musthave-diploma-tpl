package config

import "github.com/caarlos0/env"

type Config struct {
	AppPort     int    `env:"APP_PORT" envDefault:"8080"`
	DatabaseDSN string `env:"DATABASE_DSN"`
}

func New() (Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

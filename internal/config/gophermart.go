package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
)

type GophermartConfig struct {
	TokenExp    time.Duration `env:"JWT_EXP"`
	TokenSecret string        `env:"JWT_SECRET"`
}

func LoadGophermartConfig() (*GophermartConfig, error) {
	srvConf := &GophermartConfig{}

	flag.StringVar(&srvConf.TokenSecret, "sec", "secret", "jwt secret")
	flag.DurationVar(&srvConf.TokenExp, "exp", 3*time.Hour, "jwt expiration period")

	flag.Parse()

	err := env.Parse(srvConf)
	if err != nil {
		return nil, err
	}

	return srvConf, nil
}

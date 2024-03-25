package config

import (
	"flag"

	"github.com/caarlos0/env"
)

type ServerConfig struct {
	Host           string `env:"ADDRESS"`
	DSN            string `env:"DATABASE_DSN"`
	Secret         string `env:"SECRET" envDefault:"key"`
	AccuralSysAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewServerConfig() *ServerConfig {
	config := parseServerFlags()
	serverConfig := &ServerConfig{}
	env.Parse(serverConfig)

	if serverConfig.Host == "" {
		serverConfig.Host = config.Host
	}

	if serverConfig.DSN == "" {
		serverConfig.DSN = config.DSN
	}

	if serverConfig.Secret == "" {
		serverConfig.Secret = config.Secret
	}

	if serverConfig.AccuralSysAddr == "" {
		serverConfig.AccuralSysAddr = config.AccuralSysAddr
	}

	return serverConfig
}

func parseServerFlags() *ServerConfig {
	config := &ServerConfig{}

	flag.StringVar(&config.Host, "a", "localhost:8080", "server host")
	flag.StringVar(&config.DSN, "d", "user=postgres password=postgres host=127.0.0.1 port=5432 dbname=gophermartDB sslmode=disable", "DB connection string")
	flag.StringVar(&config.AccuralSysAddr, "r", "", "accural addr")

	flag.Parse()

	return config
}

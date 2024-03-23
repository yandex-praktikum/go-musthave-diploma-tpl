package config

import (
	"flag"

	"github.com/caarlos0/env"
)

type ServerConfig struct {
	Host           string `env:"ADDRESS"`
	DSN            string `env:"DATABASE_DSN"`
	LogLevel       string `env:"LOG_LEVEL"`
	StoragePath    string `env:"STORAGE_PATH"`
	Secret         string `env:"SECRET" envDefault:"key"`
	AccuralSysAddr string `env:"RESTORE"`
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

	if serverConfig.LogLevel == "" {
		serverConfig.LogLevel = config.LogLevel
	}

	if serverConfig.StoragePath == "" {
		serverConfig.StoragePath = config.StoragePath
	}

	if serverConfig.Secret == "" {
		serverConfig.Secret = config.Secret
	}

	return serverConfig
}

func parseServerFlags() *ServerConfig {
	config := &ServerConfig{}

	flag.StringVar(&config.Host, "a", "localhost:8080", "server host")
	flag.StringVar(&config.DSN, "d", "", "DB connection string")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.StringVar(&config.StoragePath, "f", "/tmp/metrics-db.json", "path to file to store metrics")
	flag.StringVar(&config.Secret, "s", "", "secret key for signing data")
	flag.StringVar(&config.AccuralSysAddr, "r", "", "store metrics in file")

	flag.Parse()

	return config
}

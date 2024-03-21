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
	Key            string `env:"KEY"`
	AccuralSysAddr string `env:"RESTORE"`
	StoreInvterval int64  `env:"STORE_INTERVAL"`
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

	if serverConfig.Key == "" {
		serverConfig.Key = config.Key
	}

	return serverConfig
}

func parseServerFlags() *ServerConfig {
	config := &ServerConfig{}

	flag.StringVar(&config.Host, "a", "localhost:8080", "server host")
	flag.StringVar(&config.DSN, "d", "", "DB connection string")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.StringVar(&config.StoragePath, "f", "/tmp/metrics-db.json", "path to file to store metrics")
	flag.StringVar(&config.Key, "k", "", "secret key for signing data")
	flag.StringVar(&config.AccuralSysAddr, "r", "", "store metrics in file")

	flag.Int64Var(&config.StoreInvterval, "i", 2, "interval of storing metrics")

	flag.Parse()

	return config
}

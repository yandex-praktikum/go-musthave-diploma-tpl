package config

import (
	"flag"
)

type ServiceConfig struct {
	ServiceDomain string
	LogLevel      string
	DatabaseDsn   string
}

func ParseConfig() ServiceConfig {
	config := ServiceConfig{}
	flag.StringVar(&config.ServiceDomain, "a", DefaultDomain, "")
	flag.StringVar(&config.LogLevel, "l", DefaultLoggerLevel, "")
	flag.StringVar(&config.DatabaseDsn, "d", DefaultDatabaseDsn, "")

	flag.Parse()
	return config
}

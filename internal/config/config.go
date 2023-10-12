package config

import (
	"flag"

	"github.com/caarlos0/env"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/constants"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/logger"
)

type ConfigServer struct {
	Port        string `env:"RUN_ADDRESS "`
	AccrualPort string `env:"ACCRUAL_SYSTEM_ADDRESS "`
	DSN         string `env:"DATABASE_URI"`
	Logger      *logger.Config
}

func InitServer() (*ConfigServer, error) {

	var flagRunAddr string
	var flagRunAddrAccrual string
	var flagDSN string

	cfg := &ConfigServer{}
	_ = env.Parse(cfg)

	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	// flag.StringVar(&flagRunAddrAccrual, "r", "localhost:8090", "address and port to run server")
	flag.StringVar(&flagRunAddrAccrual, "r", "", "address and port to run server")
	flag.StringVar(&flagDSN, "d", "sslmode=disable host=localhost port=5432 dbname = gofermart user=dbuser password=password123", "connection to database")

	flag.Parse()

	if cfg.Port == "" {
		cfg.Port = flagRunAddr
	}

	if cfg.AccrualPort == "" {
		cfg.AccrualPort = flagRunAddrAccrual
	}

	if cfg.DSN == "" {
		cfg.DSN = flagDSN
	}

	cfglog := &logger.Config{
		LogLevel: constants.LogLevel,
		DevMode:  constants.DevMode,
		Type:     constants.Type,
	}

	cfg.Logger = cfglog

	return cfg, nil
}

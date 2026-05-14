package main

import (
	"flag"
	"log"
	"strings"

	"github.com/caarlos0/env/v6"
)

type Options struct {
	Addr                 string `env:"RUN_ADDRESS"`
	DatabaseDSN          string `env:"DATABASE_DSN"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`

	// Дополнительная настройка через config.yml
	ConfigPath string
}

func parseFlag(opt *Options) {
	opt.Addr = "localhost:8080"
	flag.Func("a", "server endpoint", func(s string) error {
		parts := strings.Split(s, ":")
		if len(parts) == 3 {
			opt.Addr = strings.Trim(parts[1], "/") + ":" + parts[2]
			return nil
		}
		opt.Addr = s
		return nil
	})

	opt.AccrualSystemAddress = "localhost:8081"
	flag.Func("r", "server endpoint", func(s string) error {
		parts := strings.Split(s, ":")
		if len(parts) == 3 {
			opt.AccrualSystemAddress = strings.Trim(parts[1], "/") + ":" + parts[2]
			return nil
		}
		opt.AccrualSystemAddress = s
		return nil
	})

	flag.StringVar(&opt.DatabaseDSN, "d", "", "database dsn")
	flag.StringVar(&opt.ConfigPath, "config_path", "Path of config", "The path of config file")

	flag.Parse()

	err := env.Parse(opt)
	if err != nil {
		log.Fatalf("failed to parse: %s", err.Error())
	}
}

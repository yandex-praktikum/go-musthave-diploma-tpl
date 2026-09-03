// Package config reads application settings from defaults, flags and env.
package config

import (
	"flag"
	"fmt"
	"os"
)

const (
	defaultRunAddress           = ":8080"
	defaultDatabaseURI          = ""
	defaultAccrualSystemAddress = ""
)

// Config contains application runtime settings.
type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

// Load reads configuration from command-line flags and environment variables.
func Load() (Config, error) {
	return Parse(os.Args[1:], os.LookupEnv)
}

// Parse reads configuration from args and lookup, where lookup usually is os.LookupEnv.
func Parse(args []string, lookup func(string) (string, bool)) (Config, error) {
	cfg := Config{
		RunAddress:           defaultRunAddress,
		DatabaseURI:          defaultDatabaseURI,
		AccrualSystemAddress: defaultAccrualSystemAddress,
	}

	fs := flag.NewFlagSet("gophermart", flag.ContinueOnError)
	fs.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "HTTP server address")
	fs.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "PostgreSQL connection URI")
	fs.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "accrual system address")
	if err := fs.Parse(args); err != nil {
		return Config{}, fmt.Errorf("parse flags: %w", err)
	}

	if value, ok := lookup("RUN_ADDRESS"); ok {
		cfg.RunAddress = value
	}
	if value, ok := lookup("DATABASE_URI"); ok {
		cfg.DatabaseURI = value
	}
	if value, ok := lookup("ACCRUAL_SYSTEM_ADDRESS"); ok {
		cfg.AccrualSystemAddress = value
	}

	return cfg, nil
}

package main

import (
	"flag"
	"os"
)

type Config struct {
	endpoint        string
	accrualEndpoint string
	dsn             string
	logLevel        string
	env             string
}

func NewConfig() Config {
	var (
		endpoint        string
		accrualEndpoint string
		dsn             string
		logLevel        string
		env             string
	)

	flag.StringVar(&endpoint, "a", "localhost:8090", "address and port to run server")
	flag.StringVar(&accrualEndpoint, "r", "localhost:8080", "address and port to accrual run server")
	flag.StringVar(&dsn, "d", "", "data source name for database connection")
	flag.Parse()

	if address := os.Getenv("RUN_ADDRESS"); address != "" {
		endpoint = address
	}

	if accrualAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accrualAddress != "" {
		accrualEndpoint = accrualAddress
	}

	if d := os.Getenv("DATABASE_URI"); d != "" {
		dsn = d
	}

	if l := os.Getenv("LOG_LEVEL"); l != "" {
		logLevel = l
	} else {
		logLevel = "error"
	}

	if e := os.Getenv("ENV"); e != "" {
		env = e
	} else {
		env = "production"
	}

	return Config{
		endpoint,
		accrualEndpoint,
		dsn,
		logLevel,
		env,
	}
}

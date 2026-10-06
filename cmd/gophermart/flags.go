package main

import (
	"flag"
	"fmt"
)

var flagLogLevel string

var flagDSN string

var flagJWTSecretKey string

var flagAccrualAddress string

var flagServiceAddress string

func parseFlags() {
	flag.StringVar(&flagLogLevel, "l", "DEBUG", "log level (DEBUG, INFO, WARN, ERROR)")
	flag.StringVar(&flagDSN, "d", "", "database connection DSN")
	flag.StringVar(&flagJWTSecretKey, "k", "", "secret key used to sign JWT tokens")
	flag.StringVar(&flagAccrualAddress, "r", "127.0.0.1:8080", "accrual system address")
	flag.StringVar(&flagServiceAddress, "a", ":8888", "service address and port to listen on")
	flag.Parse()
}

func validateFlags() error {
	if flagDSN == "" {
		return fmt.Errorf("database URI is not provided")
	}
	if flagLogLevel == "" {
		return fmt.Errorf("log level is not provided")
	}
	return nil
}

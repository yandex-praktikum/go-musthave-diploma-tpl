package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
)

var (
	argumentLogLevel       string = "DEBUG"
	argumentDSN            string
	argumentJWTSecretKey   string
	argumentAccrualAddress string = "http://127.0.0.1:8080"
	argumentServiceAddress string = "127.0.0.1:8888"
)

func parseEnvs() {
	envLogLevel, found := os.LookupEnv("LOG_LEVEL")
	if found {
		argumentLogLevel = envLogLevel
	}

	envDSN, found := os.LookupEnv("DATABASE_URI")
	if found {
		argumentDSN = envDSN
	}

	envJWTSecretKey, found := os.LookupEnv("JWT_SECRET_KEY")
	if found {
		argumentJWTSecretKey = envJWTSecretKey
	}

	envAccrualAddress, found := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if found {
		argumentAccrualAddress = envAccrualAddress
	}

	envServiceAddress, found := os.LookupEnv("RUN_ADDRESS")
	if found {
		argumentServiceAddress = envServiceAddress
	}
}

func parseFlags() {
	flag.StringVar(&argumentLogLevel, "l", argumentLogLevel, "log level (DEBUG, INFO, WARN, ERROR)")
	flag.StringVar(&argumentDSN, "d", argumentDSN, "database connection DSN")
	flag.StringVar(&argumentJWTSecretKey, "k", argumentJWTSecretKey, "secret key used to sign JWT tokens")
	flag.StringVar(&argumentAccrualAddress, "r", argumentAccrualAddress, "accrual system URL, e.g. http://127.0.0.1:8080")
	flag.StringVar(&argumentServiceAddress, "a", argumentServiceAddress, "loyalty service address and port to listen on")
	flag.Parse()
}

func validateArguments() error {
	if argumentLogLevel == "" {
		return fmt.Errorf("log level is not provided")
	}
	if argumentDSN == "" {
		return fmt.Errorf("database URI is not provided")
	}
	u, err := url.Parse(argumentAccrualAddress)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("accrual system address must include scheme and host, e.g. http://127.0.0.1:8080")
	}
	if argumentServiceAddress == "" {
		return fmt.Errorf("loyalty service address and port is not provided")
	}
	return nil
}

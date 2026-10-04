package main

import (
	"os"
)

func parseEnvs() error {
	envLogLevel, found := os.LookupEnv("LOG_LEVEL")
	if found {
		flagLogLevel = envLogLevel
	}

	envDSN, found := os.LookupEnv("DATABASE_DSN")
	if found {
		flagDSN = envDSN
	}

	envJWTSecretKey, found := os.LookupEnv("JWT_SECRET_KEY")
	if found {
		flagJWTSecretKey = envJWTSecretKey
	}

	envServiceAddress, found := os.LookupEnv("RUN_ADDRESS")
	if found {
		flagServiceAddress = envServiceAddress
	}

	envAccrualAddress, found := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if found {
		flagAccrualAddress = envAccrualAddress
	}

	return nil
}

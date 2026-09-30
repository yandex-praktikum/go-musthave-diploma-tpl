package main

import (
	"fmt"
	"os"
	"strconv"
)

func parseEnvs() error {
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

	envQueueCapacity, found := os.LookupEnv("QUEUE_CAPACITY")
	if found {
		queueCapacity, err := strconv.Atoi(envQueueCapacity)
		if err != nil {
			return fmt.Errorf("failed to parse QUEUE_CAPACITY: %w", err)
		}

		flagQueueCapacity = queueCapacity
	}

	return nil
}

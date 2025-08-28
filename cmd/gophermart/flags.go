package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type flags struct {
	serverAddr           string
	postgresDSN          string
	accrualSystemAddress string
}

func initFlags() (flags, error) {
	serverAddr := flag.String("a", ":8081", "The address to bind the server to")
	postgresDSN := flag.String("d", "", "The flag to Postgres DSN")
	accrualSystemAddress := flag.String("r", "", "The accrual system address")

	flag.Parse()

	if value := os.Getenv("RUN_ADDRESS"); value != "" {
		serverAddr = &value
	}

	dataBaseDSNKey := "DATABASE_URI"
	if value, exist := os.LookupEnv(dataBaseDSNKey); exist {
		if value == "" {
			return flags{}, fmt.Errorf("%s environment variable not set", dataBaseDSNKey)
		}

		postgresDSN = &value
	}

	accrualSystemAddressKey := "ACCRUAL_SYSTEM_ADDRESS"
	if value, exist := os.LookupEnv(accrualSystemAddressKey); exist {
		if value == "" {
			return flags{}, fmt.Errorf("%s environment variable not set", accrualSystemAddressKey)
		}

		accrualSystemAddress = &value
	}

	return flags{
		serverAddr:           *serverAddr,
		accrualSystemAddress: *accrualSystemAddress,
		postgresDSN:          *postgresDSN,
	}, nil
}

func parseIntervalValue(value string) (int64, error) {
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s: %w", value, err)
	}
	if intValue <= 0 {
		return 0, fmt.Errorf("invalid POLL_INTERVAL: %s", value)
	}

	return intValue, nil
}

package main

import (
	"flag"
	"fmt"
)

var flagDSN string

var flagMigrationPath string

var flagJWTSecretKey string

var flagAccrualAddress string

var flagServiceAddress string

var flagQueueCapacity int

func parseFlags() {
	flag.StringVar(&flagDSN, "d", "", "database connection DSN")
	flag.StringVar(&flagMigrationPath, "m", "./migrations", "directory containing database migration files")
	flag.StringVar(&flagJWTSecretKey, "k", "", "secret key used to sign JWT tokens")
	flag.StringVar(&flagAccrualAddress, "r", ":8080", "accrual system address")
	flag.StringVar(&flagServiceAddress, "a", ":8888", "service address and port to listen on")
	flag.IntVar(&flagQueueCapacity, "q", 20, "TODO")
	flag.Parse()
}

func validateFlags() error {
	if flagDSN == "" {
		return fmt.Errorf("DSN is not provided")
	}
	return nil
}

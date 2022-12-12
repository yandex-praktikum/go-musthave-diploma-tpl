package apiserver

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	RunAddress           string
	AccrualSystemAddress string
	LogLevel             string
	DatabaseURI          string
}

func NewConfig() *Config {
	var runAddress, accrualSystemAddress, databaseURI string
	flag.StringVar(&runAddress, "a", ":8080", "Input server address")
	flag.StringVar(&accrualSystemAddress, "r", ":8081", "Input loyality server address")
	flag.StringVar(&databaseURI, "d", "", "Input DATABASE URI")
	flag.Parse()

	//get from env
	envRunAddress := os.Getenv("RUN_ADDRESS")
	if envRunAddress != "" {
		runAddress = envRunAddress
	}

	envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	if envAccrualSystemAddress != "" {
		accrualSystemAddress = envAccrualSystemAddress
	}

	envDatabaseURI := os.Getenv("DATABASE_URI")
	if envDatabaseURI != "" {
		databaseURI = envDatabaseURI
	}

	if databaseURI == "" {
		log.Fatal("DATABASE_URI is empty")
	}

	return &Config{
		RunAddress:           runAddress,
		AccrualSystemAddress: accrualSystemAddress,
		DatabaseURI:          databaseURI,
		LogLevel:             "debug",
	}
}

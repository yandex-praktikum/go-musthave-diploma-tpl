package apiserver

import (
	"flag"
)

type Config struct {
	RunAddress           string
	AccrualSystemAddress string
	LogLevel             string
}

func NewConfig() *Config {
	var runAddress, accrualSystemAddress string
	flag.StringVar(&runAddress, "a", ":8080", "Input server address")
	flag.StringVar(&accrualSystemAddress, "r", ":8081", "Input loyality server address")
	flag.Parse()

	return &Config{
		RunAddress:           runAddress,
		AccrualSystemAddress: accrualSystemAddress,
		LogLevel:             "debug",
	}
}

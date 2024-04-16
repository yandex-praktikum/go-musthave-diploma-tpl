package main

import (
	"fmt"
	"github.com/caarlos0/env/v10"
	"github.com/evgfitil/gophermart.git/internal/logger"
	"github.com/spf13/cobra"
)

const (
	defaultRunAddress = "localhost:8080"
)

var (
	cfg     *Config
	rootCmd = &cobra.Command{
		Use:   "server",
		Short: "Gophermart Loyalty System",
		Long: `Gophermart Loyalty System is a comprehensive server-side application designed to manage a rewards-based loyalty program. 
                This system allows registered users to submit order numbers, tracks these submissions, 
                and interfaces with an external accrual system to calculate loyalty points based on user purchases.`,
		Run: runServer,
	}
)

func runServer(cmd *cobra.Command, args []string) {
	logger.InitLogger(cfg.LogLevel)
	defer logger.Sugar.Sync()

	if err := env.Parse(cfg); err != nil {
		logger.Sugar.Fatalf("error parsing config: %v", err)
	}

	/*
			!!! TO-DO
		    1. add router and run http-server here
		    2. remove this stub
	*/
	fmt.Println("Starting server...")
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cfg = NewConfig()
	rootCmd.Flags().StringVarP(&cfg.RunAddress, "address", "a", defaultRunAddress, "run address for the server in the format host:port")
	rootCmd.Flags().StringVarP(&cfg.DatabaseURI, "database-uri", "d", "", "database connection string")
	rootCmd.Flags().StringVarP(&cfg.AccrualSystemAddress, "accrual-system-address", "r", "", "accrual system address")
}

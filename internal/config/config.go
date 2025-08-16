package config

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

const (
	RUN_ADDRESS_KEY            = "RUN_ADDRESS"
	DATABASE_URI_KEY           = "DATABASE_URI"
	ACCRUAL_SYSTEM_ADDRESS_KEY = "ACCRUAL_SYSTEM_ADDRESS"
)

type Config struct {
	RunAddressValue           string
	DatabaseURIValue          string
	AccrualSystemAddressValue string
}

func NewConfig() *Config {
	return &Config{}
}

func ReadFlags(c Config) error {
	rootCmd := &cobra.Command{
		Use:   "go-shop",
		Short: "[-a address], [-d database], [-r accrual system]",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("address: %s, database: %s, accrual system: %s\n", c.RunAddressValue, c.DatabaseURIValue, c.AccrualSystemAddressValue)
		},
	}
	rootCmd.Flags().StringVarP(&c.RunAddressValue, "Port for service", "a", ":8080", "Server address")
	rootCmd.Flags().StringVarP(&c.DatabaseURIValue, "URI for Postgres DB", "d", "postgres://admin:admin@localhost/go-shop?sslmode=disable", "Postgres URI")
	rootCmd.Flags().StringVarP(&c.AccrualSystemAddressValue, "ACCRUAL SYSTEM ADDRESS", "r", ":8000", "ACCRUAL_SYSTEM_ADDRESS")

	err := rootCmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func Init() (Config, error) {

	c := NewConfig()
	c.RunAddressValue = os.Getenv(RUN_ADDRESS_KEY)
	c.DatabaseURIValue = os.Getenv(DATABASE_URI_KEY)
	c.AccrualSystemAddressValue = os.Getenv(ACCRUAL_SYSTEM_ADDRESS_KEY)

	if c.RunAddressValue == "" || c.DatabaseURIValue == "" || c.AccrualSystemAddressValue == "" {
		if err := ReadFlags(*c); err != nil {
			return Config{}, err
		}
	}
	return *c, nil
}

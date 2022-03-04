package configs

import (
	"flag"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	ServerAddress  string
	DatabaseURI    string
	AccrualAddress string
}

func InitConfig() (*Config, error) {
	//viper init
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	//env init
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	var serverAddress, dbURI, accrualAdress string

	serverAddressDefault := os.Getenv("RUN_ADDRESS")
	dbURIDefault := os.Getenv("DATABASE_URI")
	accrualAddressDefault := os.Getenv("ACCRUAL_SYSTEM_ADDRESS")

	flag.StringVar(&serverAddress, "a", serverAddressDefault, "address of API server")
	flag.StringVar(&dbURI, "d", dbURIDefault, "str to DB connection")
	flag.StringVar(&accrualAdress, "r", accrualAddressDefault, "address of accrual system")

	flag.Parse()

	return &Config{
		ServerAddress:  serverAddress,
		DatabaseURI:    dbURI,
		AccrualAddress: accrualAdress,
	}, nil
}

func NewConfigForTest() *Config {
	return &Config{
		ServerAddress:  "localhost:8080",
		AccrualAddress: "http://localhost:8080",
	}
}

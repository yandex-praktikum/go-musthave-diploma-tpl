package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string `env:"RUN_ADDRESS" env-default:":8081"`
	Accrual       string `env:"ACCRUAL_SYSTEM_ADDRESS" env-default:""`
	DB            string `env:"DATABASE_URI" env-default:""`
	SecretKey     string `env:"SECRET_KEY" env-default:"123EW"`
}

func ParseFlags() *Config {
	config := &Config{
		ServerAddress: ":8081",
		Accrual:       "http://localhost:8080/",
		DB:            "",
		SecretKey:     "123EW",
	}

	if serverAddress := os.Getenv("SERVER_ADDRESS"); serverAddress != "" {
		config.ServerAddress = serverAddress
	}

	if accrual := os.Getenv("BASE_URL"); accrual != "" {
		config.Accrual = accrual
	}

	if db := os.Getenv("DATABASE_URI"); db != "" {
		config.DB = db
	}

	if secret := os.Getenv("SECRET_KEY"); secret != "" {
		config.SecretKey = secret
	}

	serverAddressFlag := flag.String("a", "", "Server address")
	dbFlag := flag.String("d", "", "Database")
	accrualFlag := flag.String("r", "", "Accrual System address")
	secretFlag := flag.String("s", "", "Secret key")
	flag.Parse()

	if *serverAddressFlag != "" {
		config.ServerAddress = *serverAddressFlag
	}

	if *dbFlag != "" {
		config.DB = *dbFlag
	}

	if *accrualFlag != "" {
		config.Accrual = *accrualFlag
	}

	if *secretFlag != "" {
		config.SecretKey = *secretFlag
	}

	return config
}

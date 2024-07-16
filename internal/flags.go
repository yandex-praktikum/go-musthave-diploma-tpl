package internal

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"log"
	"net/url"
	"strings"
)

func InitConfig() {
	serverAddress := flag.String("a", "", "listen address")
	baseURL := flag.String("b", "", "base url")
	databaseDSN := flag.String("d", "", "database dsn")

	flag.Parse()

	if err := env.Parse(&Config); err != nil {
		fmt.Printf("%+v\n", err)
	}
	if Config.ServerAddress == "" {
		Config.ServerAddress = *serverAddress
	}
	if Config.BaseURL == "" {
		Config.BaseURL = *baseURL
	}

	if Config.ServerAddress == "" {
		Config.ServerAddress = ":8080"
	}
	if Config.BaseURL == "" {
		Config.BaseURL = "http://localhost:8080"
	}

	_, err := url.Parse(Config.BaseURL)
	if err != nil {
		log.Fatal("invalid base url: ", err)
	}

	if Config.DatabaseDSN == "" {
		Config.DatabaseDSN = *databaseDSN
	}

	Config.DatabaseDSN = strings.TrimSpace(Config.DatabaseDSN)

	if Config.JwtSecret == "" {
		Config.JwtSecret = "secret"
	}
}

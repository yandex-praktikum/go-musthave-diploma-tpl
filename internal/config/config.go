package config

import (
	"flag"
	"net/url"
	"os"
)

const defaultDataBaseAddress = "host=localhost port=5435 user=postgres password=1234 dbname=postgres sslmode=disable"
const defaultHost = "localhost"
const defaultPort = "8080"

type Config struct {
	BaseURL         string
	DataBaseAddress string
}

var configuration *Config

func GetConfig() *Config {
	if configuration == nil {
		var conf = Config{}

		flag.StringVar(&conf.BaseURL, "a", defaultHost+":"+defaultPort, "RUN_ADDRESS")
		flag.StringVar(&conf.DataBaseAddress, "d", defaultDataBaseAddress, "DATABASE_URI")
		flag.Parse()

		if envServerAddress := os.Getenv("RUN_ADDRESS"); envServerAddress != "" {
			conf.BaseURL = envServerAddress
		}

		if envDatabaseStoragePath := os.Getenv("DATABASE_URI"); envDatabaseStoragePath != "" {
			conf.DataBaseAddress = envDatabaseStoragePath
		}

		configuration = &Config{
			BaseURL:         URLParseHelper(conf.BaseURL),
			DataBaseAddress: conf.DataBaseAddress,
		}
	}

	return configuration

}

func URLParseHelper(str string) string {
	parsedURL, err := url.ParseRequestURI(str)
	if err != nil {
		return ""
	}

	if parsedURL.Host == "" {
		parsedURL, err = url.ParseRequestURI("http://" + str)
		if err != nil {
			return ""
		}
	}

	port := parsedURL.Port()
	if port == "" {
		port = defaultPort
	}

	host := parsedURL.Hostname()
	if host == "" {
		host = defaultHost
	}

	return host + ":" + port
}

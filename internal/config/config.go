package config

import (
	"errors"
	"flag"
	"regexp"
	"strings"

	"github.com/caarlos0/env/v6"
)

type ServerAddress string

func (address *ServerAddress) String() string {
	return string(*address)
}

func (address *ServerAddress) Set(flagValue string) error {
	if len(flagValue) == 0 {
		return errors.New("empty address not valid")
	}

	reg := regexp.MustCompile(`^([0-9A-Za-z\.]+)?(\:[0-9]+)?$`)

	if !reg.MatchString(flagValue) {
		return errors.New("invalid address and port")
	}

	*address = ServerAddress(flagValue)
	return nil
}

type Config struct {
	Server        ServerAddress `env:"RUN_ADDRESS"`
	DatabaseDsn   string        `env:"DATABASE_URI"`
	AccrualServer ServerAddress `env:"ACCRUAL_SYSTEM_ADDRESS"`
	SecretKey     string        `env:"KEY"`
}

var config = Config{
	Server:        ":8080",
	AccrualServer: ":8081",
	DatabaseDsn:   "",
	SecretKey:     "",
}

func init() {
	flag.Var(&config.Server, "a", "address and port to run server")
	flag.StringVar(&config.DatabaseDsn, "d", "", "connection string for postgre")
	flag.Var(&config.AccrualServer, "r", "address and port to connect an accrual server")
	flag.StringVar(&config.SecretKey, "k", "", "sha256 based secret key")
}

func MustLoad() *Config {
	flag.Parse()

	err := env.Parse(&config)

	transformServerAddress(&config.AccrualServer)

	if err != nil {
		panic(err)
	}

	return &config
}

// resty нужен протокол, в тестах указывается без протокола, обходим
func transformServerAddress(address *ServerAddress) {
	if !strings.HasPrefix(address.String(), "https://") && !strings.HasPrefix(address.String(), "http://") {
		*address = ServerAddress("http://" + address.String())
	}
}

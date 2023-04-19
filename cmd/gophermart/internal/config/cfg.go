package config

import (
	"flag"
	"os"
	"time"
)

const (
	dataBaseURI = "postgres://qwerty:qwerty@localhost:5436/postgres?sslmode=disable"

	timeLive          = 30 * time.Minute
	accessTokenSecret = "key_gophermarket_secret"
	userKey           = "user_id"
	repetitionClient  = 3
)

type Config struct {
	ServerAddress         string
	AccrualSystemAddress  string
	DataBaseURI           string
	CountRepetitionBD     string
	CountRepetitionClient int
	Token                 TokenConfig
}

type TokenConfig struct {
	SecretKey           string
	AccessTimeLiveToken time.Duration
	UserKey             string
}

func New() *Config {
	cfg := new(Config)
	cfg.Token = TokenConfig{
		SecretKey:           accessTokenSecret,
		AccessTimeLiveToken: timeLive,
		UserKey:             userKey,
	}

	if cfg.ServerAddress = os.Getenv("RUN_ADDRESS"); cfg.ServerAddress == "" {
		flag.StringVar(&cfg.ServerAddress, "a", ":8080", "Server address")
	}

	if cfg.AccrualSystemAddress = os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); cfg.AccrualSystemAddress == "" {
		flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Accural system address")
	}

	if cfg.DataBaseURI = os.Getenv("DATABASE_URI"); cfg.DataBaseURI == "" {
		flag.StringVar(&cfg.DataBaseURI, "d", dataBaseURI, "")
	}

	if cfg.CountRepetitionBD = os.Getenv("REPETITION_CONNECT"); cfg.CountRepetitionBD == "" {
		flag.StringVar(&cfg.CountRepetitionBD, "repetition", "5", "repetition connect database")
	}

	cfg.CountRepetitionClient = repetitionClient

	flag.Parse()
	return cfg
}

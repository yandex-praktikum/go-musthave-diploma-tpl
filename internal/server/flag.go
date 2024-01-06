package server

import (
	"flag"
	"fmt"
	"log"

	"github.com/caarlos0/env/v6"
)

var HPServer *string
var ServerKey *string
var PsqlInfo *string
var ASAdress *string

type ServerEnvConfig struct {
	Address              *string `env:"RUN_ADDRESS"`
	ConnectionDBString   *string `env:"DATABASE_URI"`
	KeyForHash           *string `env:"KEY"`
	AccrualSystemAddress *string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func ParseArgsServer() {
	var cfg ServerEnvConfig
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	HPServer = flag.String("a", ":8080", "host and port in format <host>:<port>")
	PsqlInfo = flag.String(
		"d",
		"host=localhost port=5432 user=main password=620631 dbname=bonussystem sslmode=disable",
		"Connection string for psql",
	)
	ServerKey = flag.String("k", "asdsfavasxc", "Key to hash passwords")
	ASAdress = flag.String("r", ":8080", "Accrual system adress. Host and port format")
	flag.Parse()
	if cfg.Address != nil {
		HPServer = cfg.Address
	}
	if cfg.AccrualSystemAddress != nil {
		ASAdress = cfg.AccrualSystemAddress
	}
	if cfg.KeyForHash != nil {
		ServerKey = cfg.KeyForHash
	}
	if cfg.ConnectionDBString != nil {
		PsqlInfo = cfg.ConnectionDBString
	}
	fmt.Println("Connection string for psql:", *PsqlInfo)
	fmt.Println("Connection string for accrual system:", *ASAdress)
	fmt.Println("Server key to hash/auth:", *ServerKey)
	fmt.Println("Adress of server:", *HPServer)
}

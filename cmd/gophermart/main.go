package main

import (
	"log"

	"github.com/brisk84/gofemart/internal/app"
	"github.com/brisk84/gofemart/internal/config"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	a, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(a.Run())
}

package main

import (
	"github.com/EestiChameleon/GOphermart/internal/app/cfg"
	"github.com/EestiChameleon/GOphermart/internal/app/router"
	"github.com/EestiChameleon/GOphermart/internal/app/storage"
	"log"
)

func main() {
	if err := cfg.GetEnvs(); err != nil {
		log.Fatal(err)
	}

	if err := storage.InitConnection(); err != nil {
		log.Fatal(err)
	}
	defer storage.Shutdown()

	if err := router.Start(); err != nil {
		log.Fatal(err)
	}
}

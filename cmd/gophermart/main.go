package main

import (
	"github.com/EestiChameleon/BonusServiceGOphermart/internal/app/cfg"
	"github.com/EestiChameleon/BonusServiceGOphermart/internal/app/router"
	"github.com/EestiChameleon/BonusServiceGOphermart/internal/app/storage"
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

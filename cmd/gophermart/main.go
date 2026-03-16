package main

import (
	"log"

	"gophermart/internal/config"
	"gophermart/internal/server"
)

func main() {
	cfg := config.Load()

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("create server: %v", err)
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}

package main

import (
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/config"
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/logger"
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/server"
)

func main() {
	cfg, err := config.InitServer()
	if err != nil {

		panic("error initialazing config")
	}

	appLogger := logger.NewAppLogger(cfg.Logger)
	appLogger.InitLogger()

	srv := server.NewServer(cfg, appLogger)

	if err := srv.Run(); err != nil {
		panic(err)
	}
}

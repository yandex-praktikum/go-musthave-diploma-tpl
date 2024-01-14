package main

import (
	"github.com/k-morozov/go-musthave-diploma-tpl/app"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/config"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/logger"
	"github.com/k-morozov/go-musthave-diploma-tpl/servers"
)

func main() {
	cfg := config.ParseConfig()
	log := logger.NewLogger(cfg.LogLevel)

	application, err := app.NewApp(cfg,
		log,
		servers.OptLogger(log),
		servers.WithAuth())

	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	application.Run()
}

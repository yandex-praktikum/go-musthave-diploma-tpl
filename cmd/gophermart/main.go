package main

import (
	"encoding/json"
	server "github.com/GTech1256/go-musthave-diploma-tpl/internal"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/config"
	"github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	"log"
)

func main() {
	cfg := config.NewConfig()
	err := cfg.Load()
	if err != nil {
		log.Fatalln(err)
		return
	}

	logging.Init()
	logger := logging.NewMyLogger()

	cfgMarshal, err := json.MarshalIndent(&cfg, "", "   ")
	if err == nil {
		logger.Info("CFG: ", string(cfgMarshal))
	}

	logger.Info("Запуск приложения")
	_, err = server.New(cfg, logger)
	if err != nil {
		logger.Error("Запуск приложения провалено", err)
		return

	}
}

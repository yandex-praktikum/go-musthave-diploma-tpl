package main

import (
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/app"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/config"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/logger"
	"go.uber.org/zap"
)

func main() {
	AppConfig := config.New()

	config.ParseEnv(AppConfig)
	config.ParseFlags(AppConfig)

	if err := app.Run(AppConfig); err != nil {
		panic(err)
	}
	logger.Log.Info("running server", zap.String("address", AppConfig.RunAddress))
}

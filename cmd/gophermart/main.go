package main

import (
	"log"

	"github.com/NailUsmanov/internal/app"
	"github.com/NailUsmanov/internal/storage"
	"github.com/NailUsmanov/pkg/config"
	"go.uber.org/zap"
)

func main() {

	// Cоздаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// Создаем регистратор SugaredLogger
	sugar := logger.Sugar()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbStorage, err := storage.NewDataBaseStorage(cfg.DataBaseURI)
	if err != nil {
		sugar.Fatalf("failed to connect to database: %v", err)
	}

	applictaion := app.NewApp(dbStorage, sugar, cfg.Accural)
	if err := applictaion.Run(cfg.RunAddr); err != nil {
		sugar.Fatalln(err)
	}

}

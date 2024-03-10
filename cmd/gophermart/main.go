package main

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"

	"github.com/A-Kuklin/gophermart/internal/config"
	"github.com/A-Kuklin/gophermart/internal/storage/postgres"
)

func main() {
	cfg := config.LoadConfig()
	logrus.SetLevel(cfg.LogLevel)

	strg, err := postgres.NewStorage(cfg.DBdsn)
	if err != nil {
		logrus.Errorf("Error while creating PSQL storage: %s", err)
	}

	err = strg.MigrateUP(cfg)
	if err != nil {
		logrus.Errorf("Error while up migration: %s", err)
	}
	logrus.Info("Migrations were finished")

}

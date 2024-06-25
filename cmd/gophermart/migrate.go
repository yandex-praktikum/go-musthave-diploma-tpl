package main

import (
	"database/sql"
	"github.com/ShukinDmitriy/gophermart/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	"os"
	"path"
)

func runMigrate(e *echo.Echo, config config.Config) error {
	db, err := sql.Open("postgres", config.DatabaseURI)
	if err != nil {
		e.Logger.Error("can't connect to db", err.Error())
		return err
	}
	defer func() {
		db.Close()
	}()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		e.Logger.Error("can't create driver", err.Error())
		return err
	}

	currentDir, _ := os.Getwd()
	m, err := migrate.NewWithDatabaseInstance(
		"file:///"+path.Join(currentDir, "db", "migrations"),
		"postgres", driver)
	if err != nil {
		e.Logger.Error("can't create new migrate: ", err.Error())
		return err
	}

	err = m.Up()
	if err != nil {
		e.Logger.Info("can't migrate up: ", err.Error())
	}

	return nil
}

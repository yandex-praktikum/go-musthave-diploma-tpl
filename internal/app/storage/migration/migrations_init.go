package migration

import (
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var (
	m *migrate.Migrate
)

func UpGophermartStorage() error {
	err := m.Up()
	if !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func DownGophermartStorage() error {
	return m.Down()
}

func MigrateInitConnect() error {
	// как подключаться к яндекс базе???
	conn, err := sql.Open("postgres",
		"user=maximiliank password='' dbname=yandex_practicum_db sslmode=disable")
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		return err
	}

	db, err := migrate.NewWithDatabaseInstance(
		"file://internal/app/storage/migration/sqlscripts/",
		"postgres", driver)
	if err != nil {
		return err
	}

	m = db
	return nil
}

func MigrateCloseConnect() (error, error) {
	return m.Close()
}

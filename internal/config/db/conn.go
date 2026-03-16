package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//var embedMigrations embed.FS

func NewConnection(ctx context.Context, ps string) (*sql.DB, error) {

	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии базы данных %w", err)
	}

	err = migrations(ps)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func migrations(ps string) error {
	m, err := migrate.New(
		"file://migrations", //migrations
		ps)
	if err != nil {
		return fmt.Errorf("ошибка создания объекта миграции: %w  Строка подключения - %s", err, ps)
	}

	// Применение миграций
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

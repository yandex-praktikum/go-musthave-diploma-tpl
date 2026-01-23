package db

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewConnection(ctx context.Context, ps string) (*sql.DB, error) {

	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии базы данных %w", err)
	}

	err = runMigrations(ctx, db)
	if err != nil {
		return nil, err
	}

	return db, nil
}
func runMigrations(ctx context.Context, db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	} //filepath.Join("..", "migrations")
	//var embedMigrations embed.FS
	//goose.SetBaseFS(embedMigrations)

	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

func migrations(ps string) error {

	// Запуск миграций при старте приложения
	m, err := migrate.New(
		"file:"+filepath.Join("..", "migrations"),
		ps)
	if err != nil {
		//log.Fatalf("Ошибка создания объекта миграции: %v", err)
		//log.Println("Ошибка создания объекта миграции:  " + err.Error() + ". Строка подключения - " + ps)
		return fmt.Errorf("Ошибка создания объекта миграции: %w. Строка подключения - %s", err, ps)
	}

	// Применение миграций
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

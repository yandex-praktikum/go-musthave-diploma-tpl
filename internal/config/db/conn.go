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

//func runMigrations(ctx context.Context, db *sql.DB) error {
//	if err := goose.SetDialect("postgres"); err != nil {
//		return err
//	}
//	goose.SetLogger(goose.NopLogger())
//	goose.SetBaseFS(embedMigrations)
//
//	migrationsDir := "migrations"
//
//	log.Printf("applying goose migrations from %s", migrationsDir)
//	if err := goose.UpContext(ctx, db, migrationsDir); err != nil {
//		return fmt.Errorf("goose up: %w", err)
//	}
//	return nil
//}

func migrations(ps string) error {
	fmt.Println("миграте")
	// Запуск миграций при старте приложения
	m, err := migrate.New(
		"file:migrations",
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

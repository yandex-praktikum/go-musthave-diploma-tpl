package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewConnection(ps string) (*sql.DB, error) {

	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии базы данных %w", err)
	}
	err = migration(ps)
	if err != nil {
		return nil, err
	}

	return db, nil
}
func migration(ps string) error {
	// Запуск миграций при старте приложения
	m, err := migrate.New(
		"file://../migrations",
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

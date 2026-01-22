package db

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewConnection(ps string) (*sql.DB, error) {

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
func runMigrations(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, getMigrationsDir()) // путь работает нормально
}
func getMigrationsDir() string {
	// Получаем путь к текущему файлу (main.go)
	////_, filename, _, _ := runtime.Caller(0)
	// Берём директорию: .../cmd/gophermart
	//gophermartDir := filepath.Dir(filename)
	// Поднимаемся на уровень: .../cmd
	//cmdDir := filepath.Dir(gophermartDir)
	// Спускаемся в migrations
	return filepath.Join("..", "migrations")
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

package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Storage представляет хранилище данных с репозиториями
type Storage struct {
	Pool    *pgxpool.Pool
	Users   *UserRepository
	Orders  *OrderRepository
	Balance *BalanceRepository
}

// New создаёт новое хранилище с подключением к базе данных
func New(ctx context.Context, dsn string) (*Storage, error) {
	// Создаем пул подключений
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	// Проверяем подключение
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Применяем миграции
	if err := runMigrations(pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	// Создаем репозитории
	storage := &Storage{
		Pool:    pool,
		Users:   NewUserRepository(pool),
		Orders:  NewOrderRepository(pool),
		Balance: NewBalanceRepository(pool),
	}

	return storage, nil
}

// runMigrations применяет миграции базы данных
func runMigrations(pool *pgxpool.Pool) error {
	// Для golang-migrate нужен стандартный sql.DB
	// Получаем connection string и создаем sql.DB
	sqlDB, err := sql.Open("postgres", pool.Config().ConnString())
	if err != nil {
		return fmt.Errorf("open sql connection: %w", err)
	}
	defer sqlDB.Close()

	// Создаем драйвер базы данных для migrate
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres driver: %w", err)
	}

	// Создаем источник миграций из embed FS
	source, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("create iofs source: %w", err)
	}

	// Создаем экземпляр migrate
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	// Применяем миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

// Close закрывает подключение к базе данных
func (s *Storage) Close() {
	if s.Pool != nil {
		s.Pool.Close()
	}
}

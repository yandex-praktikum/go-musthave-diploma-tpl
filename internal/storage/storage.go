package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// StorageOption — функция для настройки Storage.
type StorageOption func(*storageConfig)

type storageConfig struct {
	migrationsPath string
	maxOpenConns   int
	maxIdleConns   int
}

// WithMigrationsPath устанавливает путь к миграциям.
func WithMigrationsPath(path string) StorageOption {
	return func(c *storageConfig) {
		c.migrationsPath = path
	}
}

// WithMaxOpenConns устанавливает максимальное количество открытых соединений.
func WithMaxOpenConns(n int) StorageOption {
	return func(c *storageConfig) {
		c.maxOpenConns = n
	}
}

// WithMaxIdleConns устанавливает максимальное количество простаивающих соединений.
func WithMaxIdleConns(n int) StorageOption {
	return func(c *storageConfig) {
		c.maxIdleConns = n
	}
}

// Storage — подключение к БД и репозитории.
type Storage struct {
	db                  *sql.DB
	userRepository      repository.UserRepository
	orderRepository     repository.OrderRepository
	withdrawalRepository repository.WithdrawalRepository
}

// UserRepository возвращает репозиторий пользователей.
func (s *Storage) UserRepository() repository.UserRepository {
	return s.userRepository
}

// OrderRepository возвращает репозиторий заказов.
func (s *Storage) OrderRepository() repository.OrderRepository {
	return s.orderRepository
}

// WithdrawalRepository возвращает репозиторий списаний.
func (s *Storage) WithdrawalRepository() repository.WithdrawalRepository {
	return s.withdrawalRepository
}

// InitializeStorage запускает миграции, подключается к БД и создаёт репозитории.
// При пустом dsn возвращает ошибку.
func InitializeStorage(dsn string, opts ...StorageOption) (*Storage, error) {
	if dsn == "" {
		return nil, errors.New("database DSN required")
	}
	
	cfg := &storageConfig{
		migrationsPath: "",
		maxOpenConns:   0,
		maxIdleConns:   0,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	
	if err := RunMigrations(dsn, cfg.migrationsPath); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	
	if cfg.maxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.maxOpenConns)
	}
	if cfg.maxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.maxIdleConns)
	}
	
	return &Storage{
		db:                   db,
		userRepository:       postgres.NewUserRepository(db),
		orderRepository:      postgres.NewOrderRepository(db),
		withdrawalRepository: postgres.NewWithdrawalRepository(db),
	}, nil
}

// Close закрывает подключение к БД.
func (s *Storage) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

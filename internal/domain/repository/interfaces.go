package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
)

// ==================== Константы ====================

const (
	pingTimeout = 5 * time.Second
)

// ==================== Ошибки ====================

var (
	ErrDuplicateKey        = errors.New("duplicate key violation")
	ErrForeignKey          = errors.New("foreign key violation")
	ErrConnectionFailed    = errors.New("database connection failed")
	ErrNotFound            = errors.New("not found")
	ErrUserNotActive       = errors.New("user not active")
	ErrUserHasActiveOrders = errors.New("user has active orders")
	ErrUserAlreadyExists   = errors.New("user already exists")
)

// ==================== Интерфейсы репозиториев ====================

// Store основной интерфейс для работы с хранилищем
type Store interface {
	User() UserRepository
	Order() OrderRepository
	Withdrawal() WithdrawalRepository

	Ping(ctx context.Context) error
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	Close()

	HealthCheck(ctx context.Context) error
}

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, login, password string) (*entity.User, error)
	GetByID(ctx context.Context, id int) (*entity.User, error)
	GetByLogin(ctx context.Context, login string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id int) error
	UpdatePassword(ctx context.Context, userID int, newPassword string) error
	Deactivate(ctx context.Context, userID int) error
	GetBalance(ctx context.Context, userID int) (*entity.UserBalance, error)

	Exists(ctx context.Context, userID int) (bool, error)
	GetStats(ctx context.Context, userID int) (*entity.UserStats, error)
	GetCount(ctx context.Context) (int, error)
	GetAll(ctx context.Context, limit, offset int) ([]entity.User, error)
}

// OrderRepository интерфейс для работы с заказами
type OrderRepository interface {
	Create(ctx context.Context, userID int, number string, status entity.OrderStatus) (*entity.Order, error)
	GetByID(ctx context.Context, id int) (*entity.Order, error)
	GetByNumber(ctx context.Context, number string) (*entity.Order, error)
	GetByUserID(ctx context.Context, userID int) ([]entity.Order, error)
	GetAll(ctx context.Context, limit, offset int) ([]entity.OrderWithUser, error)
	Update(ctx context.Context, order *entity.Order) error
	UpdateStatus(ctx context.Context, orderID int, status entity.OrderStatus) error
	UpdateAccrual(ctx context.Context, orderID int, accrual float64, status entity.OrderStatus) error // Исправленная сигнатура
	Delete(ctx context.Context, id int) error
	DeleteByNumber(ctx context.Context, number string) error

	GetPending(ctx context.Context, limit int) ([]entity.Order, error)
	GetByStatus(ctx context.Context, status entity.OrderStatus) ([]entity.Order, error)
	Exists(ctx context.Context, number string) (bool, error)
	GetForProcessing(ctx context.Context, limit int) ([]entity.Order, error)
	GetActive(ctx context.Context) ([]entity.Order, error)
	UpdateStatusByNumber(ctx context.Context, number string, status entity.OrderStatus, accrual *float64) error
}

// WithdrawalRepository интерфейс для работы со списаниями
type WithdrawalRepository interface {
	Create(ctx context.Context, withdrawal *entity.Withdrawal) error
	GetByUserID(ctx context.Context, userID int, status entity.OrderStatus) ([]entity.Withdrawal, error)
	GetAll(ctx context.Context, limit, offset int) ([]entity.Withdrawal, error)
	GetTotalWithdrawn(ctx context.Context, userID int) (float64, error)
}

// ==================== Реализация репозитория ====================

// Repository реализация хранилища с использованием PostgreSQL
type Repository struct {
	db *pgxpool.Pool
}

// userRepo встроенная реализация UserRepository
type userRepo struct {
	db *pgxpool.Pool
}

// orderRepo встроенная реализация OrderRepository
type orderRepo struct {
	db *pgxpool.Pool
}

// withdrawalRepo встроенная реализация WithdrawalRepository
type withdrawalRepo struct {
	db *pgxpool.Pool
}

// NewRepository создает новый экземпляр репозитория
func NewRepository(pool *pgxpool.Pool) (*Repository, error) {
	if pool == nil {
		return nil, fmt.Errorf("%w: pool is nil", ErrConnectionFailed)
	}

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	return &Repository{db: pool}, nil
}

// ==================== Методы доступа к репозиториям ====================

// User возвращает репозиторий для работы с пользователями
func (r *Repository) User() UserRepository {
	return &userRepo{db: r.db}
}

// Order возвращает репозиторий для работы с заказами
func (r *Repository) Order() OrderRepository {
	return &orderRepo{db: r.db}
}

// Withdrawal возвращает репозиторий для работы со списаниями
func (r *Repository) Withdrawal() WithdrawalRepository {
	return &withdrawalRepo{db: r.db}
}

// ==================== Управление соединением и транзакциями ====================

// Ping проверяет соединение с базой данных
func (r *Repository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// HealthCheck выполняет расширенную проверку здоровья базы данных
func (r *Repository) HealthCheck(ctx context.Context) error {
	// Проверяем соединение
	if err := r.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Проверяем доступность ключевых таблиц
	tables := []string{"users", "orders"}
	for _, table := range tables {
		query := fmt.Sprintf("SELECT 1 FROM %s LIMIT 1", table)
		if _, err := r.db.Exec(ctx, query); err != nil {
			return fmt.Errorf("table %s is not accessible: %w", table, err)
		}
	}

	return nil
}

// BeginTx начинает новую транзакцию
func (r *Repository) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return r.db.BeginTx(ctx, opts)
}

// Close закрывает все соединения с базой данных
func (r *Repository) Close() {
	if r.db != nil {
		r.db.Close()
	}
}

// ==================== Вспомогательные функции ====================

// Scanner интерфейс для сканирования данных из базы
type Scanner interface {
	Scan(dest ...interface{}) error
}

// IsDuplicateError проверяет, является ли ошибка нарушением уникальности
func IsDuplicateError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}
	return false
}

// IsForeignKeyError проверяет, является ли ошибка нарушением внешнего ключа
func IsForeignKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503" // foreign_key_violation
	}
	return false
}

// IsNotNullError проверяет, является ли ошибка нарушением ограничения NOT NULL
func IsNotNullError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23502" // not_null_violation
	}
	return false
}

// IsCheckError проверяет, является ли ошибка нарушением ограничения CHECK
func IsCheckError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23514" // check_violation
	}
	return false
}

// IsConnectionError проверяет, является ли ошибка ошибкой соединения
func IsConnectionError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// Коды ошибок соединения в PostgreSQL
		connectionErrorCodes := []string{
			"08000", // connection_exception
			"08001", // sqlclient_unable_to_establish_sqlconnection
			"08003", // connection_does_not_exist
			"08004", // sqlserver_rejected_establishment_of_sqlconnection
			"08006", // connection_failure
			"08007", // transaction_resolution_unknown
		}
		for _, code := range connectionErrorCodes {
			if pgErr.Code == code {
				return true
			}
		}
	}
	return false
}

// WrapError оборачивает ошибку с дополнительным контекстом
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}

	if IsDuplicateError(err) {
		return fmt.Errorf("%w: %s", ErrDuplicateKey, context)
	}

	if IsForeignKeyError(err) {
		return fmt.Errorf("%w: %s", ErrForeignKey, context)
	}

	return fmt.Errorf("%s: %w", context, err)
}

package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store интерфейс репозитория
type Store interface {
	UsersRepository
	OrdersRepository
	Ping(ctx context.Context) error
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	Close()
}

// UsersRepository интерфейс для работы с пользователями
type UsersRepository interface {
	CreateUser(ctx context.Context, login, password string) (*entity.User, error)
	GetUserByID(ctx context.Context, id int) (*entity.User, error)
	GetUserByLogin(ctx context.Context, login string) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, id int) error
	UpdateUserPassword(ctx context.Context, userID int, newPassword string) error
	DeactivateUser(ctx context.Context, userID int) error
	GetUserBalance(ctx context.Context, userID int) (*entity.UserBalance, error)
}

// OrdersRepository интерфейс для работы с заказами
type OrdersRepository interface {
	CreateOrder(ctx context.Context, userID int, number string) (*entity.Order, error)
	GetOrderByID(ctx context.Context, id int) (*entity.Order, error)
	GetOrderByNumber(ctx context.Context, number string) (*entity.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int) ([]entity.Order, error)
	GetAllOrders(ctx context.Context, limit, offset int) ([]entity.OrderWithUser, error)
	UpdateOrder(ctx context.Context, order *entity.Order) error
	UpdateOrderStatus(ctx context.Context, orderID int, status entity.OrderStatus) error
	UpdateOrderAccrual(ctx context.Context, orderID int, accrual float64) error
	DeleteOrder(ctx context.Context, id int) error
	GetPendingOrders(ctx context.Context, limit int) ([]entity.Order, error)
	GetOrdersByStatus(ctx context.Context, status entity.OrderStatus) ([]entity.Order, error)
	ExistsByNumber(ctx context.Context, number string) (bool, error)
	GetOrdersForProcessing(ctx context.Context, limit int) ([]*entity.Order, error)
	GetActiveOrders(ctx context.Context) ([]*entity.Order, error)
	UpdateStatus(ctx context.Context, number string, status entity.OrderStatus, accrual *float64) error
	CreateWithdraw(ctx context.Context, withdraw *entity.Withdraw) error
}

// Repository реализация Store
type Repository struct {
	db *pgxpool.Pool
}

func (r *Repository) DeleteOrder(ctx context.Context, id int) error {
	//TODO implement me
	panic("implement me")
}

// NewRepository создает новый репозиторий
func NewRepository(pool *pgxpool.Pool) (*Repository, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Repository{db: pool}, nil
}

// Ping проверяет соединение с БД
func (r *Repository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// BeginTx начинает транзакцию
func (r *Repository) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return r.db.BeginTx(ctx, opts)
}

// Close закрывает соединение с БД
func (r *Repository) Close() {
	r.db.Close()
}

// IsDuplicateError проверяет, является ли ошибка нарушением уникальности
func IsDuplicateError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// IsForeignKeyError проверяет, является ли ошибка нарушением внешнего ключа
func IsForeignKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}

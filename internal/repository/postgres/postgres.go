// Package postgres реализует репозиторий для работы с PostgreSQL
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"iter"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"

	"github.com/anon-d/gophermarket/internal/repository"
	"github.com/anon-d/gophermarket/migrations"
)


// PostgresDB репозиторий для работы с PostgreSQL
type PostgresDB struct {
	db         *sqlx.DB
	logger     *zap.Logger
	userRepo   *GenericRepository[repository.User]
	orderRepo  *GenericRepository[repository.Order]
	balanceRepo *GenericRepository[repository.Balance]
	withdrawalRepo *GenericRepository[repository.Withdrawal]
}

// NewPostgres создаёт новый экземпляр репозитория
func NewPostgres(dsn string, logger *zap.Logger) (*PostgresDB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	repo := &PostgresDB{
		db:             db,
		logger:         logger,
		userRepo:       NewGenericRepository[repository.User](db),
		orderRepo:      NewGenericRepository[repository.Order](db),
		balanceRepo:    NewGenericRepository[repository.Balance](db),
		withdrawalRepo: NewGenericRepository[repository.Withdrawal](db),
	}

	// Run migrations
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := repo.migrate(ctx); err != nil {
		return nil, fmt.Errorf("failed to run migrations in NewRepository: %w", err)
	}

	return repo, nil
}

// Close закрывает соединение с БД
func (p *PostgresDB) Close() error {
	return p.db.Close()
}

func (p *PostgresDB) migrate(ctx context.Context) error {
	goose.SetBaseFS(migrations.Migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect in migrate: %w", err)
	}

	if err := goose.UpContext(ctx, p.db.DB, "."); err != nil {
		return fmt.Errorf("failed to apply migrations in migrate: %w", err)
	}

	p.logger.Info("Migrations applied successfully")
	return nil
}

// CreateUser создаёт нового пользователя
func (p *PostgresDB) CreateUser(ctx context.Context, user *repository.User) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Проверяем, существует ли пользователь
	var exists bool
	err = tx.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM users WHERE login = $1)", user.Login)
	if err != nil {
		return err
	}
	if exists {
		return repository.ErrUserExists
	}

	// Создаём пользователя
	_, err = tx.ExecContext(ctx,
		"INSERT INTO users (id, login, pass_hash) VALUES ($1, $2, $3)",
		user.ID, user.Login, user.PassHash)
	if err != nil {
		return err
	}

	// Создаём баланс для пользователя
	_, err = tx.ExecContext(ctx,
		"INSERT INTO balances (user_id, current_balance, withdrawn) VALUES ($1, 0, 0)",
		user.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetUser получает пользователя по ID
func (p *PostgresDB) GetUser(ctx context.Context, uid string) (*repository.User, error) {
	return p.userRepo.GetOne(ctx, "SELECT id, login, pass_hash FROM users WHERE id = $1", uid)
}

// GetUserByLogin получает пользователя по логину
func (p *PostgresDB) GetUserByLogin(ctx context.Context, login string) (*repository.User, error) {
	return p.userRepo.GetOne(ctx, "SELECT id, login, pass_hash FROM users WHERE login = $1", login)
}

// CreateOrder создаёт новый заказ
func (p *PostgresDB) CreateOrder(ctx context.Context, order *repository.Order) error {
	// Проверяем, существует ли заказ
	var existingOrder repository.Order
	err := p.db.GetContext(ctx, &existingOrder,
		"SELECT id, number, user_id, status, accrual, uploaded_at FROM orders WHERE number = $1",
		order.Number)
	if err == nil {
		// Заказ существует
		if existingOrder.UserID == order.UserID {
			return repository.ErrOrderExists
		}
		return repository.ErrOrderExistsByAnotherUser
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: %w", repository.ErrInternal, err)
	}

	_, err = p.db.ExecContext(ctx,
		"INSERT INTO orders (number, user_id, status, accrual, uploaded_at) VALUES ($1, $2, $3, $4, $5)",
		order.Number, order.UserID, order.Status, order.Accrual, order.UploadedAt)
	return err
}

// GetOrdersByUserID получает заказы пользователя
func (p *PostgresDB) GetOrdersByUserID(ctx context.Context, userID string) ([]repository.Order, error) {
	return p.orderRepo.GetMany(ctx, "SELECT id, number, user_id, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC", userID)
}

// GetOrdersForProcessing получает заказы для обработки (NEW или PROCESSING)
func (p *PostgresDB) GetOrdersForProcessing(ctx context.Context) ([]repository.Order, error) {
	var orders []repository.Order
	err := p.db.SelectContext(ctx, &orders,
		"SELECT id, number, user_id, status, accrual, uploaded_at FROM orders WHERE status IN ('NEW', 'PROCESSING') ORDER BY uploaded_at ASC")
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// StreamOrdersForProcessing возвращает итератор для заказов в обработке
func (p *PostgresDB) StreamOrdersForProcessing(ctx context.Context) iter.Seq[repository.Order] {
	return func(yield func(repository.Order) bool) {
		orders, err := p.GetOrdersForProcessing(ctx)
		if err != nil {
			p.logger.Error("Ошибка получения заказов для обработки", zap.Error(err))
			return
		}

		for _, order := range orders {
			if !yield(order) {
				return
			}
		}
	}
}

// UpdateOrderStatus обновляет статус заказа
func (p *PostgresDB) UpdateOrderStatus(ctx context.Context, orderNumber string, status string, accrual float64) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Обновляем статус заказа
	_, err = tx.ExecContext(ctx,
		"UPDATE orders SET status = $1, accrual = $2 WHERE number = $3",
		status, accrual, orderNumber)
	if err != nil {
		return err
	}

	// Если статус PROCESSED и есть начисление, обновляем баланс
	if status == "PROCESSED" && accrual > 0 {
		// Получаем user_id заказа
		var userID uuid.UUID
		err = tx.GetContext(ctx, &userID, "SELECT user_id FROM orders WHERE number = $1", orderNumber)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx,
			"UPDATE balances SET current_balance = current_balance + $1 WHERE user_id = $2",
			accrual, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetBalance получает баланс пользователя
func (p *PostgresDB) GetBalance(ctx context.Context, userID string) (*repository.Balance, error) {
	var balance repository.Balance
	err := p.db.GetContext(ctx, &balance,
		"SELECT user_id, current_balance, withdrawn FROM balances WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// Withdraw списывает баллы со счёта пользователя
func (p *PostgresDB) Withdraw(ctx context.Context, withdrawal *repository.Withdrawal) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Проверяем баланс
	var currentBalance float64
	err = tx.GetContext(ctx, &currentBalance,
		"SELECT current_balance FROM balances WHERE user_id = $1 FOR UPDATE",
		withdrawal.UserID)
	if err != nil {
		return err
	}

	if currentBalance < withdrawal.Sum {
		return repository.ErrInsufficientFunds
	}

	// Списываем с баланса
	_, err = tx.ExecContext(ctx,
		"UPDATE balances SET current_balance = current_balance - $1, withdrawn = withdrawn + $1 WHERE user_id = $2",
		withdrawal.Sum, withdrawal.UserID)
	if err != nil {
		return err
	}

	// Создаём запись о списании
	_, err = tx.ExecContext(ctx,
		"INSERT INTO withdrawals (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, $4)",
		withdrawal.UserID, withdrawal.OrderNumber, withdrawal.Sum, withdrawal.ProcessedAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetWithdrawals получает список списаний пользователя
func (p *PostgresDB) GetWithdrawals(ctx context.Context, userID string) ([]repository.Withdrawal, error) {
	return p.withdrawalRepo.GetMany(ctx, "SELECT id, user_id, order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC", userID)
}

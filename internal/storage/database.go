package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/models"
)

// DatabaseStorage реализация хранилища на PostgreSQL
type DatabaseStorage struct {
	pool *pgxpool.Pool
}

// NewDatabaseStorage создает новое подключение к базе данных
func NewDatabaseStorage(ctx context.Context, databaseURI string) (*DatabaseStorage, error) {
	pool, err := pgxpool.New(ctx, databaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseStorage{pool: pool}, nil
}

// Close закрывает соединение с базой данных
func (s *DatabaseStorage) Close() error {
	s.pool.Close()
	return nil
}

// Ping проверяет соединение с базой данных
func (s *DatabaseStorage) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// CreateUser создает нового пользователя
func (s *DatabaseStorage) CreateUser(ctx context.Context, login, passwordHash string) (*models.User, error) {
	var user models.User
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, login, password_hash`

	err := s.pool.QueryRow(ctx, query, login, passwordHash).Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// GetUserByLogin получает пользователя по логину
func (s *DatabaseStorage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	query := `SELECT id, login, password_hash FROM users WHERE login = $1`

	err := s.pool.QueryRow(ctx, query, login).Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}

	return &user, nil
}

// GetUserByID получает пользователя по ID
func (s *DatabaseStorage) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	query := `SELECT id, login, password_hash FROM users WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// CreateOrder создает новый заказ
func (s *DatabaseStorage) CreateOrder(ctx context.Context, userID int64, number string) (*models.Order, error) {
	var order models.Order
	query := `INSERT INTO orders (user_id, number, status, uploaded_at) VALUES ($1, $2, $3, $4) RETURNING id, user_id, number, status, accrual, uploaded_at`

	now := time.Now()
	err := s.pool.QueryRow(ctx, query, userID, number, "NEW", now).Scan(
		&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return &order, nil
}

// GetOrderByNumber получает заказ по номеру
func (s *DatabaseStorage) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	var order models.Order
	query := `SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE number = $1`

	err := s.pool.QueryRow(ctx, query, number).Scan(
		&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order by number: %w", err)
	}

	return &order, nil
}

// GetOrdersByUserID получает все заказы пользователя
func (s *DatabaseStorage) GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by user id: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// UpdateOrderStatus обновляет статус заказа
func (s *DatabaseStorage) UpdateOrderStatus(ctx context.Context, number string, status string, accrual *float64) error {
	query := `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`

	_, err := s.pool.Exec(ctx, query, status, accrual, number)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

// GetBalance получает баланс пользователя
func (s *DatabaseStorage) GetBalance(ctx context.Context, userID int64) (*models.Balance, error) {
	var balance models.Balance
	query := `SELECT user_id, current, withdrawn FROM balances WHERE user_id = $1`

	err := s.pool.QueryRow(ctx, query, userID).Scan(&balance.UserID, &balance.Current, &balance.Withdrawn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Создаем новый баланс если не существует
			balance = models.Balance{UserID: userID, Current: 0, Withdrawn: 0}
			err = s.UpdateBalance(ctx, userID, 0, 0)
			if err != nil {
				return nil, fmt.Errorf("failed to create balance: %w", err)
			}
			return &balance, nil
		}
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &balance, nil
}

// UpdateBalance обновляет баланс пользователя
func (s *DatabaseStorage) UpdateBalance(ctx context.Context, userID int64, current, withdrawn float64) error {
	query := `INSERT INTO balances (user_id, current, withdrawn) VALUES ($1, $2, $3) 
			  ON CONFLICT (user_id) DO UPDATE SET current = $2, withdrawn = $3`

	_, err := s.pool.Exec(ctx, query, userID, current, withdrawn)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return nil
}

// CreateWithdrawal создает новое списание средств
func (s *DatabaseStorage) CreateWithdrawal(ctx context.Context, userID int64, order string, sum float64) (*models.Withdrawal, error) {
	var withdrawal models.Withdrawal
	query := `INSERT INTO withdrawals (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, $4) RETURNING id, user_id, order_number, sum, processed_at`

	now := time.Now()
	err := s.pool.QueryRow(ctx, query, userID, order, sum, now).Scan(
		&withdrawal.ID, &withdrawal.UserID, &withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create withdrawal: %w", err)
	}

	return &withdrawal, nil
}

// GetWithdrawalsByUserID получает все списания пользователя
func (s *DatabaseStorage) GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	query := `SELECT id, user_id, order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals by user id: %w", err)
	}
	defer rows.Close()

	var withdrawals []models.Withdrawal
	for rows.Next() {
		var withdrawal models.Withdrawal
		err := rows.Scan(&withdrawal.ID, &withdrawal.UserID, &withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	return withdrawals, nil
}

// GetOrdersByStatus получает заказы по статусам
func (s *DatabaseStorage) GetOrdersByStatus(ctx context.Context, statuses []string) ([]models.Order, error) {
	if len(statuses) == 0 {
		return []models.Order{}, nil
	}

	// Строим запрос с параметрами для статусов
	query := `SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE status = ANY($1) ORDER BY uploaded_at ASC`

	rows, err := s.pool.Query(ctx, query, statuses)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by status: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// GetOrdersByStatusPaginated получает заказы по статусам с пагинацией
func (s *DatabaseStorage) GetOrdersByStatusPaginated(ctx context.Context, statuses []string, limit, offset int) ([]models.Order, error) {
	if len(statuses) == 0 {
		return []models.Order{}, nil
	}

	// Строим запрос с параметрами для статусов и пагинацией
	query := `SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE status = ANY($1) ORDER BY uploaded_at ASC LIMIT $2 OFFSET $3`

	rows, err := s.pool.Query(ctx, query, statuses, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by status with pagination: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// ProcessWithdrawal выполняет списание средств в транзакции
func (s *DatabaseStorage) ProcessWithdrawal(ctx context.Context, userID int64, order string, sum float64) (*models.Withdrawal, error) {
	// Начинаем транзакцию
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Получаем баланс с блокировкой строки
	var balance models.Balance
	balanceQuery := `SELECT user_id, current, withdrawn FROM balances WHERE user_id = $1 FOR UPDATE`

	err = tx.QueryRow(ctx, balanceQuery, userID).Scan(&balance.UserID, &balance.Current, &balance.Withdrawn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Создаем новый баланс если не существует
			balance = models.Balance{UserID: userID, Current: 0, Withdrawn: 0}
		} else {
			return nil, fmt.Errorf("failed to get balance for update: %w", err)
		}
	}

	// Проверяем достаточность средств
	if balance.Current < sum {
		return nil, fmt.Errorf("insufficient funds: current balance %.2f, requested %.2f", balance.Current, sum)
	}

	// Создаем списание
	var withdrawal models.Withdrawal
	withdrawalQuery := `INSERT INTO withdrawals (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, $4) RETURNING id, user_id, order_number, sum, processed_at`

	now := time.Now()
	err = tx.QueryRow(ctx, withdrawalQuery, userID, order, sum, now).Scan(
		&withdrawal.ID, &withdrawal.UserID, &withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create withdrawal: %w", err)
	}

	// Обновляем баланс
	newCurrent := balance.Current - sum
	newWithdrawn := balance.Withdrawn + sum
	updateBalanceQuery := `INSERT INTO balances (user_id, current, withdrawn) VALUES ($1, $2, $3) 
						  ON CONFLICT (user_id) DO UPDATE SET current = $2, withdrawn = $3`

	_, err = tx.Exec(ctx, updateBalanceQuery, userID, newCurrent, newWithdrawn)
	if err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	// Подтверждаем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &withdrawal, nil
}

// UpdateOrderStatusAndBalance атомарно обновляет статус заказа и баланс пользователя
func (s *DatabaseStorage) UpdateOrderStatusAndBalance(ctx context.Context, orderNumber string, status string, accrual *float64, userID int64, newCurrent, withdrawn float64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Обновляем статус заказа
	orderQuery := `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`
	_, err = tx.Exec(ctx, orderQuery, status, accrual, orderNumber)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Обновляем баланс
	balanceQuery := `INSERT INTO balances (user_id, current, withdrawn) VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET current = $2, withdrawn = $3`
	_, err = tx.Exec(ctx, balanceQuery, userID, newCurrent, withdrawn)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

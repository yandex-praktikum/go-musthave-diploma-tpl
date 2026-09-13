package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
)

// OrderRepository — репозиторий заказов, работающий с таблицей orders.
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository создаёт новый экземпляр OrderRepository поверх переданного соединения с БД.
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create создает новый заказ
func (r *OrderRepository) Create(order *domain.Order) error {
	query := `
        INSERT INTO orders (number, user_id, status, uploaded_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `

	_, err := r.db.Exec(query,
		order.Number,
		order.UserID,
		order.Status,
		order.UploadedAt,
		order.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			// Проверяем, кто владелец заказа
			existing, findErr := r.FindByNumber(order.Number)
			if findErr != nil {
				return fmt.Errorf("failed to check existing order: %w", findErr)
			}
			if existing.UserID == order.UserID {
				return domain.ErrOrderAlreadyUploadedByUser
			}
			return domain.ErrOrderAlreadyUploadedByAnotherUser
		}
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// FindByNumber находит заказ по номеру.
//
// Поле accrual может быть NULL в БД — в этом случае order.Accrual остаётся nil,
// и в JSON ответа поле не попадает (omitempty).
func (r *OrderRepository) FindByNumber(number string) (*domain.Order, error) {
	order := &domain.Order{}
	query := `
        SELECT number, user_id, status, accrual, uploaded_at, updated_at
        FROM orders
        WHERE number = $1
    `

	var accrual sql.NullFloat64
	err := r.db.QueryRow(query, number).Scan(
		&order.Number,
		&order.UserID,
		&order.Status,
		&accrual,
		&order.UploadedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	if accrual.Valid {
		order.Accrual = &accrual.Float64
	}

	return order, nil
}

// FindByUserID находит все заказы пользователя, сортирует по времени загрузки (новые сначала).
//
// Поле accrual может быть NULL в БД — в этом случае order.Accrual остаётся nil.
func (r *OrderRepository) FindByUserID(userID int64) ([]*domain.Order, error) {
	query := `
        SELECT number, user_id, status, accrual, uploaded_at, updated_at
        FROM orders
        WHERE user_id = $1
        ORDER BY uploaded_at DESC
    `

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		order := &domain.Order{}
		var accrual sql.NullFloat64

		err := rows.Scan(
			&order.Number,
			&order.UserID,
			&order.Status,
			&accrual,
			&order.UploadedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if accrual.Valid {
			order.Accrual = &accrual.Float64
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate orders: %w", err)
	}

	return orders, nil
}

// Update обновляет заказ
func (r *OrderRepository) Update(order *domain.Order) error {
	query := `
        UPDATE orders
        SET status = $1, accrual = $2, updated_at = $3
        WHERE number = $4
    `

	var accrual interface{} = nil
	if order.Accrual != nil {
		accrual = *order.Accrual
	}

	_, err := r.db.Exec(query, order.Status, accrual, order.UpdatedAt, order.Number)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

// GetPendingOrders возвращает заказы, ожидающие обработки (NEW или PROCESSING).
//
// Поле accrual может быть NULL в БД — в этом случае order.Accrual остаётся nil.
func (r *OrderRepository) GetPendingOrders(limit int) ([]*domain.Order, error) {
	query := `
        SELECT number, user_id, status, accrual, uploaded_at, updated_at
        FROM orders
        WHERE status IN ($1, $2)
        ORDER BY uploaded_at ASC
        LIMIT $3
    `

	rows, err := r.db.Query(query, domain.OrderStatusNew, domain.OrderStatusProcessing, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		order := &domain.Order{}
		var accrual sql.NullFloat64

		err := rows.Scan(
			&order.Number,
			&order.UserID,
			&order.Status,
			&accrual,
			&order.UploadedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending order: %w", err)
		}

		if accrual.Valid {
			order.Accrual = &accrual.Float64
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate pending orders: %w", err)
	}

	return orders, nil
}

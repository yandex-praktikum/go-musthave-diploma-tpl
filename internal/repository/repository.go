package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/AnxVit/go-musthave-diploma-tpl/internal/logger"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/model"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/utils"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

// Сохрание пользователя
func (r *Repository) SaveUser(ctx context.Context, login, hash_password string) (string, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx,
		"INSERT INTO users(login, hash_password) VALUES ($1, $2) RETURNING id",
		login, hash_password).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", ErrorLoginAlreadyExists
		}
		return "", err
	}
	return id.String(), err
}

// Получение id и хэш-пароля пользователя
func (r *Repository) GetPasswordByEmail(ctx context.Context, login string) (string, string, error) {
	var (
		hash_password string
		id            uuid.UUID
	)
	err := r.pool.QueryRow(ctx, "SELECT id, hash_password FROM users WHERE login = $1", login).Scan(
		&id,
		&hash_password,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", ErrorNoUser
		}
		return "", "", err
	}

	return id.String(), hash_password, nil
}

// Сохранение order
func (r *Repository) SaveOrder(ctx context.Context, order string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			logger.Log.Warn("Couldn't rollback", zap.Error(err))
		}
	}()

	currentUserID := utils.GetUserID(ctx)

	var userID uuid.UUID
	err = tx.QueryRow(ctx, `SELECT user_id FROM orders WHERE order_number = $1`, order).Scan(&userID)
	if err == nil {
		if currentUserID == userID.String() {
			return ErrorOrderAlreadyExists
		}
		return ErrorOrderConflict
	}

	cmdTag, err := tx.Exec(ctx, `INSERT INTO orders(user_id, order_number) VALUES ($1, $2)`, currentUserID, order)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("couldn't insert order")
	}

	return tx.Commit(ctx)
}

// Получение всех order пользователя
func (r *Repository) GetOrders(ctx context.Context, userID string) ([]model.Order, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT 
			order_number, 
			status, 
			accrual, 
			created_at 
		FROM orders 
		WHERE user_id = $1
		ORDER BY created_at DESC; 
		`, userID)
	if err != nil {
		return nil, err
	}

	result := make([]model.Order, 0)

	defer rows.Close()
	for rows.Next() {
		var order model.Order
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}

		result = append(result, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Получение баланса пользователя
func (r *Repository) GetBalance(ctx context.Context, userID string) (model.Balance, error) {
	var res model.Balance
	err := r.pool.QueryRow(ctx,
		`SELECT 
			COALESCE(SUM(accrual), 0) - COALESCE(SUM(withdrawn), 0) AS current,
			COALESCE(SUM(withdrawn), 0) AS withdrawn
		FROM orders 
		WHERE user_id = $1::UUID;`, userID).Scan(&res.Current, &res.Withdrawn)
	if err != nil {
		return res, err
	}

	return res, nil
}

// Проверка на возможность списания
// В случае возможность добавляется order на списание
func (r *Repository) SetWithdrawnOrder(
	ctx context.Context, userID, orderNumber string, withdrawn float64,
) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			logger.Log.Warn("Couldn't rollback", zap.Error(err))
		}
	}()

	var corrert bool
	err = tx.QueryRow(ctx, `
        SELECT COALESCE(SUM(accrual), 0) - COALESCE(SUM(withdrawn), 0) >= $1
        FROM orders 
        WHERE user_id = $2
    `, withdrawn, userID).Scan(&corrert)
	if err != nil {
		return err
	}
	if !corrert {
		return ErrorInsufficientBalance
	}

	cmdTag, err := tx.Exec(ctx,
		`INSERT INTO orders (user_id, order_number, withdrawn) VALUES ($1, $2, $3)`,
		userID, orderNumber, withdrawn,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrorOrderAlreadyExists
		}
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("couldn't insert withdrawn")
	}
	return tx.Commit(ctx)
}

// Получение всех списаний
func (r *Repository) SetWithdrawnsByUser(ctx context.Context, userID string) ([]model.Withdraw, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT 
			order_number, 
			withdrawn,
			created_at 
		FROM orders 
		WHERE user_id = $1 AND withdrawn IS NOT NULL`, userID)
	if err != nil {
		return nil, err
	}

	result := make([]model.Withdraw, 0)

	defer rows.Close()
	for rows.Next() {
		var order model.Withdraw
		err := rows.Scan(&order.Order, &order.Sum, &order.ProcessedAt)
		if err != nil {
			return nil, err
		}

		result = append(result, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

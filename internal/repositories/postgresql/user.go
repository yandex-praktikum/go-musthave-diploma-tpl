package postgresql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/KalessinD/gophermart/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	PsqlUserTable = `"` + PsqlGophermartSchema + `"."users"`

	QueryInsertUser = `INSERT INTO ` + PsqlUserTable + ` AS t (login, hash) VALUES($1, $2) RETURNING id`

	QuerySelectUserByLogin = `SELECT id, login, hash, version, balance, created_at FROM ` + PsqlUserTable + ` WHERE login = $1`
	QuerySelectUserByID    = `SELECT id, login, hash, version, balance, created_at FROM ` + PsqlUserTable + ` WHERE id = $1`

	QueryUpdateUserBalance = `UPDATE ` + PsqlUserTable + ` SET balance = balance + $2 WHERE id = $1 AND (balance + $2) > 0 RETURNING id`
)

func (r *SQLStorage) GetUser(ctx context.Context, login string) (*models.User, error) {
	user := &models.User{}

	_, err := r.withRetry(ctx, func(ctx context.Context) (*sql.Row, error) {
		row := r.db.QueryRowContext(ctx, QuerySelectUserByLogin, login)
		if err := row.Scan(&user.ID, &user.Login, &user.Hash, &user.Version, &user.Balance, &user.CreatedAt); err != nil {
			return nil, err
		}
		return row, nil
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *SQLStorage) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	user := &models.User{}

	_, err := r.withRetry(ctx, func(ctx context.Context) (*sql.Row, error) {
		row := r.db.QueryRowContext(ctx, QuerySelectUserByID, userID)
		if err := row.Scan(&user.ID, &user.Login, &user.Hash, &user.Version, &user.Balance, &user.CreatedAt); err != nil {
			return nil, err
		}
		return row, nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *SQLStorage) AddUser(ctx context.Context, user *models.User) error {
	return r.withTxRetry(ctx, func(tx *sql.Tx) error {
		if err := r.addUserTx(ctx, tx, user); err != nil {
			return err
		}
		return nil
	})
}

func (r *SQLStorage) addUserTx(ctx context.Context, tx *sql.Tx, user *models.User) error {
	err := tx.QueryRowContext(ctx, QueryInsertUser, user.Login, user.Hash).Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return models.ErrUserExists
			}
		}
		return err
	}
	return nil
}

func (r *SQLStorage) updateUserBalanceTx(ctx context.Context, tx *sql.Tx, userID string, diffSum int) error {
	var updatedUserID string
	err := tx.QueryRowContext(ctx, QueryUpdateUserBalance, userID, diffSum).Scan(&updatedUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ErrUserBalanceIsNotEnough
		}
		return err
	}
	if updatedUserID == "" { // на всякий случай
		return models.ErrUserBalanceIsNotEnough
	}
	return nil
}

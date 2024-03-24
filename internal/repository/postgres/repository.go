package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/SmoothWay/gophermart/internal/model"
	"github.com/google/uuid"
)

var ErrNotEnoughFunds = errors.New("not enough balance")

type Repository struct {
	l  *slog.Logger
	db *sql.DB
}

func New(connection *sql.DB, logger *slog.Logger) *Repository {
	return &Repository{
		db: connection,
		l:  logger,
	}
}

func (r *Repository) WithdrawalRequest(ctx context.Context, userID uuid.UUID, orderNumber string, amount float64) error {
	withdrawID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	if _, err := r.db.Exec(`INSERT INTO withdrawals (id, user_id, order_number, amount) VALUES ($1, $2, $3, $4)`, withdrawID, userID, orderNumber, amount); err != nil {
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var balance float64

	err = tx.QueryRow(`SELECT balance FROM balances WHERE user_id = $1`, userID).Scan(&balance)
	if err != nil {
		return err
	}

	balance -= amount

	if balance < 0 {
		return ErrNotEnoughFunds
	}

	_, err = tx.Exec(`UPDATE balances SET balance = $1 WHERE user_id = $2`, balance, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE withdraw_balances SET amount = (SELECT amount + $1 FROM withdraw_balances WHERE user_id = $2) WHERE user_id = $2`, amount, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE withdrawls SET status = $1 WHERE order_number = $2`, `SUCCESS`, orderNumber)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (r *Repository) AddOrder(ctx context.Context, userID uuid.UUID, order model.Order) error {
	orderID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, `INSERT INTO orders (id, user_id, order_number, status, accrual) VALUES ($1, $2, $3, $4, $5)`, orderID, userID, order.Number, order.Status, order.Accrual); err != nil {
		return err
	}

	if order.Status == "PROCESSED" {
		if _, err = tx.ExecContext(ctx, `UPDATE balances SET balance = (SELECT balance + $1 FROM balances WHERE user_id = $2) WHERE user_id = $2`, order.Accrual, userID); err != nil {
			return err
		}
	}

	tx.Commit()
	return nil
}

func (r *Repository) UpdateOrder(ctx context.Context, userID uuid.UUID, order model.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `UPDATE orders SET status = $1, accrual = $2, updated_at = $3 WHERE user_id = $4 AND order_number = $5`, order.Status, order.Accrual, order.UploadedAt, userID, order.Number)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE balances SET balance = (SELECT balance + $1 FROM balances WHERE user_id = $2) WHERE user_id = $2", order.Accrual, userID)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (r *Repository) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]model.Withdrawal, error) {

	rows, err := r.db.QueryContext(ctx, `SELECT order_number, sum, processed_at FROM withdrawls WHERE user_id = $1 AND status = $2`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	withdrawls := make([]model.Withdrawal, 0)
	w := model.Withdrawal{}

	for rows.Next() {
		err = rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawls = append(withdrawls, w)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return withdrawls, nil
}

func (r *Repository) GetBalance(ctx context.Context, userID uuid.UUID) (float64, float64, error) {
	var balance float64
	var withdrawn float64
	err := r.db.QueryRowContext(ctx, `SELECT balance FROM balances WHERE user_id = $1`, userID).Scan(&balance)
	if err != nil {
		return 0, 0, err
	}

	err = r.db.QueryRowContext(ctx, `SELECT sum FROM withdraw_balances WHERE user_id = $1`, userID).Scan(&withdrawn)
	if err != nil {
		return 0, 0, err
	}

	return balance, withdrawn, nil
}

func (r *Repository) GetOrder(ctx context.Context, userID uuid.UUID, orderNumber string) (*model.Order, error) {
	var number string

	err := r.db.QueryRowContext(ctx, `SELECT order_number FROM orders WHERE user_id = $1 AND order_number = $2`, userID, orderNumber).Scan(&number)
	if err != nil {
		return nil, err
	}

	return &model.Order{
		Number: number,
	}, nil
}

func (r *Repository) GetOrders(ctx context.Context, userID uuid.UUID) ([]model.Order, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT order_number, status, accrual, uploaded_at FROM orders WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	o := model.Order{}
	for rows.Next() {
		err := rows.Scan(&o.Number, &o.Status, &o.Accrual, &o.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *Repository) AddUser(ctx context.Context, login, password string) error {
	userID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT INTO users (id, login, password) VALUES ($1, $2, $3)`, userID, login, password)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO balances (user_id) VALUES ($1)`, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO withdraw_balances (user_id) VALUES ($1)`, userID)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (r *Repository) GetUser(ctx context.Context, login, password string) (*model.User, error) {
	var u model.User

	err := r.db.QueryRowContext(ctx, `SELECT id, login, password FROM users WHERE login = $1 AND password = $2`, login, password).Scan(&u.ID, &u.Login, &u.Password)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *Repository) ScanOrders(ctx context.Context) <-chan model.Order {
	ch := make(chan model.Order)
	wg := &sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-time.After(5 * time.Second):
				r.l.Info("Scanning orders")
				rows, err := r.db.QueryContext(ctx, `SELECT order_number, status, user_id FROM orders WHERE status != $1 and status != $2`, "PROCESSED", "INVALID")
				if err != nil {
					r.l.Info("Scanning orders error", slog.String("error", err.Error()))
					continue
				}
				defer rows.Close()

				var o model.Order

				for rows.Next() {
					err = rows.Scan(&o.Number, &o.Status, &o.UserID)
					if err != nil {
						r.l.Info("Scanning orders error", slog.String("error", err.Error()))
						continue
					}

					ch <- o
				}

				err = rows.Err()
				if err != nil {
					r.l.Info("Scanning orders error", slog.String("error", err.Error()))
					continue
				}
			case <-ctx.Done():
				r.l.Info("Scanning orders stopped")
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		r.l.Info("Closing channel")
		close(ch)
	}()

	return ch
}

func hash(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

package db

import (
	"context"
	"fmt"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"time"
)

func (d *DateBase) Save(query string, args ...interface{}) error {
	_, err := d.storage.ExecContext(context.Background(), query, args...)
	if err != nil {
		return customerrors.ErrNotFound
	}

	return nil
}

func (d *DateBase) SaveOrder(userID int, order string, now time.Time) error {
	rfc3339Time := now.Format(time.RFC3339)
	querty := "INSERT INTO loyalty (user_id, order_id, created_in) VALUES ($1, $2, $3)"

	err := d.Save(querty, userID, order, rfc3339Time)
	if err != nil {
		return err
	}

	return nil
}

func (d *DateBase) SaveTableUser(login, passwordHash string) error {
	query := "INSERT INTO users (login, password) VALUES ($1, $2)"

	// начинаем транзакцию
	tx, err := d.storage.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(context.Background(), query, login, passwordHash)
	if err != nil {
		tx.Rollback()
		return err
	}

	// завершаем транзакцию
	return tx.Commit()
}

func (d *DateBase) SaveTableUserAndUpdateToken(login, accessToken string) error {
	query := `
		UPDATE users 
		SET access_token = $2
		WHERE login = $1
	`
	_, err := d.storage.Exec(query, login, accessToken)
	if err != nil {
		return err
	}

	return nil
}

func (d *DateBase) SaveLoyaltyData(login, orderID string, bonus int, orderStatus string) error {
	var userID int
	queryGetUserID := `SELECT id FROM users WHERE login = $1`
	err := d.storage.QueryRow(queryGetUserID, login).Scan(&userID)
	if err != nil {
		return fmt.Errorf("error fetching user id: %v", err)
	}

	querySaveLoyalty := `
        INSERT INTO loyalty (user_id, order_id, bonus, order_status, is_deleted)
        VALUES ($1, $2, $3, $4, FALSE)
    `
	_, err = d.storage.Exec(querySaveLoyalty, userID, orderID, bonus, orderStatus)
	if err != nil {
		return fmt.Errorf("error inserting into loyalty: %v", err)
	}

	return nil
}

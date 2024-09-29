package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
)

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
	query := `UPDATE users SET access_token = $2 WHERE login = $1`
	_, err := d.storage.Exec(query, login, accessToken)
	if err != nil {
		return err
	}

	return nil
}

func (d *DateBase) GetLoginID(login string) (int, error) {
	var user int

	query := `SELECT id FROM users WHERE login = $1`

	row, err := d.Get(query, login)
	if err != nil {
		return -1, customerrors.ErrNotData
	}

	if err = row.Scan(&user); err != nil {
		return -1, customerrors.ErrNotUser
	}

	return user, nil
}

func (d *DateBase) SearchLoginByToken(accessToken string) (string, error) {
	var searchTokin string
	db := d.storage
	query := "SELECT access_token FROM users WHERE access_token = $1"

	// делаем запрос
	row := db.QueryRowContext(context.Background(), query, accessToken)

	if row != nil {
		return "", customerrors.ErrNotFound
	}

	if err := row.Scan(&searchTokin); err != nil {
		return "", err
	}

	return searchTokin, nil

}

func (d *DateBase) CheckTableUserLogin(login string) error {
	var existingLogin string
	query := `SELECT login FROM users WHERE login = $1`

	err := d.storage.QueryRow(query, login).Scan(&existingLogin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return customerrors.ErrUserAlreadyExists
}

func (d *DateBase) CheckTableUserPassword(login string) (string, bool) {
	var existingPassword string
	query := `SELECT password FROM users WHERE login = $1`

	err := d.storage.QueryRow(query, login).Scan(&existingPassword)

	if err != nil {
		return "", false
	}

	if existingPassword == "" {
		return "", false
	}
	return existingPassword, true
}

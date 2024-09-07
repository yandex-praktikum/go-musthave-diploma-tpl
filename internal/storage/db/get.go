package db

import (
	"context"
	"database/sql"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/custom_errors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/models"
	"time"
)

func (d *DateBase) Get(query string, args ...interface{}) (*sql.Row, error) {
	row := d.storage.QueryRow(query, args...)
	if row == nil {
		return nil, custom_errors.ErrNotFound
	}
	return row, nil
}

func (d *DateBase) GetUserByAccessToken(order string, login string, now time.Time) error {
	var user models.User
	var loyalty models.Loyalty

	queryUser := "SELECT id, login FROM users WHERE login = $1"
	queryLoyalty := "SELECT user_id FROM loyalty WHERE  order_id = $1"

	rowUser, err := d.Get(queryUser, login)
	if err != nil {
		return custom_errors.ErrNotFound
	}

	if err = rowUser.Scan(&user.ID, &user.Login); err != nil {
		return err
	}

	rowLoyalty, err := d.Get(queryLoyalty, order)
	if err != nil {
		return custom_errors.ErrNotFound
	}

	if err = rowLoyalty.Scan(&loyalty.UserID); err != nil {
		d.SaveOrder(user.ID, order, now)
		return nil
	}

	if loyalty.UserID != user.ID {
		return custom_errors.ErrAnotherUsersOrder
	}

	return custom_errors.ErrOrderIsAlready
}

func (d *DateBase) SearchLoginByToken(accessToken string) (string, error) {
	var searchTokin string
	db := d.storage
	query := "SELECT access_token FROM users WHERE access_token = $1"

	// делаем запрос
	row := db.QueryRowContext(context.Background(), query, accessToken)

	if row != nil {
		return "", custom_errors.ErrNotFound
	}

	if err := row.Scan(&searchTokin); err != nil {
		return "", err
	}

	return searchTokin, nil

}

package db

import (
	"context"
	"database/sql"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/models"
	"time"
)

func (d *DateBase) Get(query string, args ...interface{}) (*sql.Row, error) {
	row := d.storage.QueryRow(query, args...)
	if row == nil {
		return nil, customerrors.ErrNotFound
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
		return customerrors.ErrNotFound
	}

	if err = rowUser.Scan(&user.ID, &user.Login); err != nil {
		return err
	}

	rowLoyalty, err := d.Get(queryLoyalty, order)
	if err != nil {
		return customerrors.ErrNotFound
	}

	if err = rowLoyalty.Scan(&loyalty.UserID); err != nil {
		d.SaveOrder(user.ID, order, now)
		return nil
	}

	if loyalty.UserID != user.ID {
		return customerrors.ErrAnotherUsersOrder
	}

	return customerrors.ErrOrderIsAlready
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

func (d *DateBase) GetAllUserOrders(login string) ([]*models.OrdersUser, error) {
	var ordersUser []*models.OrdersUser
	tx, err := d.storage.Begin()
	if err != nil {
		return nil, err
	}

	// создаем запрос в базу users для получения id пользователя
	queryUser := "SELECT id FROM users WHERE login =$1"

	loginID, err := d.Get(queryUser, login)
	if err != nil {
		return nil, err
	}

	// создаем запрос в базу loyalty для получения всех заказов одного пользователя
	queryLoyalty := "SELECT order_id, order_status, bonus, created_in FROM loyalty WHERE user_id = $1 ORDER BY created_in ASC"

	rows, err := tx.QueryContext(context.Background(), queryLoyalty, loginID)
	if err != nil {
		return nil, sql.ErrNoRows
	}
	defer rows.Close()

	//Собираем все в ordersUser
	for rows.Next() {
		var orderUser models.OrdersUser

		if err = rows.Scan(&orderUser.Number, &orderUser.Status, &orderUser.Accrual, &orderUser.UploadedAt); err != nil {
			return nil, err
		}

		ordersUser = append(ordersUser, &orderUser)
	}
	if rows.Err() != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return ordersUser, nil
}

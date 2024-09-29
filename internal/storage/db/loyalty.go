package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/models"
	"time"
)

func (d *DateBase) SaveOrder(userID int, orderID string, orderStatus string, now time.Time) error {
	rfc3339Time := now.Format(time.RFC3339)
	querty := "INSERT INTO loyalty (user_id, order_id, order_status, created_in) VALUES ($1, $2, $3, $4)"

	if orderStatus == "" {
		orderStatus = string(models.NewOrder)
	}

	row, err := d.storage.Exec(querty, userID, orderID, orderStatus, rfc3339Time)
	if err != nil {
		return err
	}

	if row == nil {
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

		d.SaveOrder(user.ID, order, "NEW", now)
		return nil
	}

	if loyalty.UserID != user.ID {
		return customerrors.ErrAnotherUsersOrder
	}

	return customerrors.ErrOrderIsAlready
}

func (d *DateBase) GetAllUserOrders(login string) ([]*models.OrdersUser, error) {
	var ordersUser []*models.OrdersUser
	var userID int
	var zeroFloat = 0.0

	tx, err := d.storage.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	// создаем запрос в базу users для получения id пользователя
	queryUser := `SELECT id FROM users WHERE login = $1`

	rowUSer := tx.QueryRow(queryUser, login)
	err = rowUSer.Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Ошибка: пользователь не найден
			return nil, customerrors.ErrUserNotFound
		}
		// Другая ошибка
		return nil, err
	}

	// создаем запрос в базу loyalty для получения всех заказов одного пользователя
	queryLoyalty := "SELECT order_id AS Number, order_status AS Status, bonus AS Accrual, created_in AS UploadedAt FROM loyalty WHERE user_id = $1 ORDER BY created_in ASC"

	rows, err := tx.QueryContext(context.Background(), queryLoyalty, userID)
	if err != nil {
		return nil, customerrors.ErrNotFound
	}
	defer rows.Close()

	//Собираем все в ordersUser
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		var orderUser models.OrdersUser

		if err = rows.Scan(&orderUser.Number, &orderUser.Status, &orderUser.Accrual, &orderUser.UploadedAt); err != nil {
			return nil, err
		}

		if orderUser.Accrual == nil {
			orderUser.Accrual = &zeroFloat
		}

		ordersUser = append(ordersUser, &orderUser)
	}
	if rows.Err() != nil {
		return nil, err
	}

	tx.Commit()
	return ordersUser, nil
}

func (d *DateBase) GetBalanceUser(login string) (*models.Balance, error) {
	var current float32
	var withdraw float32

	// создаем запрос в базу users для получения id пользователя
	var userID int
	queryGetUserID := `SELECT id FROM users WHERE login = $1`
	err := d.storage.QueryRow(queryGetUserID, login).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching user id: %v", err)
	}

	// создаем запрос получения всех бонусов и сумм списания
	query := "SELECT SUM(bonus) AS Current, SUM(withdraw) FROM loyalty WHERE user_id = $1"

	row := d.storage.QueryRow(query, userID)

	if err := row.Scan(&current, &withdraw); err != nil {
		return nil, err
	}

	balance := models.Balance{
		Current:  current - withdraw,
		Withdraw: withdraw,
	}
	return &balance, nil
}

func (d *DateBase) GetWithdrawals(login string) ([]*models.Withdrawals, error) {
	var withdrawals []*models.Withdrawals
	var userID int

	tx, err := d.storage.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // Откат транзакции в случае ошибки

	// Создаем запрос в базу users для получения id пользователя
	queryUser := "SELECT id FROM users WHERE login = $1"
	rowUser := tx.QueryRow(queryUser, login)

	if err := rowUser.Scan(&userID); err != nil {
		return nil, err
	}

	// Создаем запрос для сбора информации по withdrawals
	query := "SELECT order_id AS order, withdraw AS Sum, processed_at AS ProcessedAt FROM loyalty WHERE user_id = $1 ORDER BY processed_at ASC"
	rows, err := tx.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Закрытие rows после использования

	for rows.Next() {
		withdraw := models.Withdrawals{}
		if err := rows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.ProcessedAt); err != nil {
			return nil, err
		}

		if withdraw.Sum != nil {
			withdrawals = append(withdrawals, &withdraw)
		}

	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if len(withdrawals) == 0 {
		return nil, customerrors.ErrNotData
	}

	return withdrawals, nil

}

func (d *DateBase) CheckWriteOffOfFunds(login, order string, sum float32, now time.Time) error {
	var sumBonus float32

	userID, err := d.GetLoginID(login)
	if err != nil {
		return err
	}

	queryCheckSumAccrual := "SELECT SUM(bonus) FROM loyalty WHERE user_id = $1"

	rowSum, err := d.Get(queryCheckSumAccrual, userID)

	if err != nil {
		return customerrors.ErrNotData
	}

	if err = rowSum.Scan(&sumBonus); err != nil {
		return customerrors.ErrNotBonus
	}

	if sum > sumBonus {
		return customerrors.ErrNotEnoughBonuses
	}

	// сохраняем новый ордер
	if err = d.SaveOrder(userID, order, models.NewOrder, now); err != nil {
		return err
	}

	// обновляем новыми данными
	querySave := "UPDATE loyalty SET processed_at = $1, withdraw = $2 WHERE order_id = $3"

	rfc3339Time := now.Format(time.RFC3339)

	if err = d.Save(querySave, rfc3339Time, sum, order); err != nil {
		return err
	}
	return nil
}

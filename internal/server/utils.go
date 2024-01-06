package server

import (
	"database/sql"
	"errors"
	"net/http"
	"syscall"
	"time"

	"github.com/akashipov/go-musthave-diploma-tpl/internal/server/general"
)

func GetWallet(request *http.Request, name string) (*Balance, error) {
	var b Balance
	f := func() error {
		getBalanceQuery := "SELECT current, withdrawn FROM balances WHERE name = $1"
		row := DB.QueryRowContext(
			request.Context(),
			getBalanceQuery, name,
		)
		err := row.Scan(&b.Current, &b.Withdrawn)
		return err
	}
	err := general.RetryCode(f, syscall.ECONNREFUSED)
	if err != nil {
		return nil, errors.New("Problem with assigning accrual " + err.Error())
	}
	return &b, nil
}

func ChangeAccrual(request *http.Request, reqData *WithdrawRequest, name string, tx *sql.Tx) error {
	f := func() error {
		changeBalance := "UPDATE balances SET current = current - $1, withdrawn = withdrawn + $1 WHERE name = $2"
		_, err := tx.ExecContext(
			request.Context(),
			changeBalance, reqData.Sum, name,
		)
		return err
	}
	err := general.RetryCode(f, syscall.ECONNREFUSED)
	if err != nil {
		return errors.New("Problem with assigning accrual " + err.Error())
	}
	return nil
}

func AddWithdrawToHistory(request *http.Request, order_id int, sum float64, name string, tx *sql.Tx) error {
	f := func() error {
		addHistoryQuery := "INSERT INTO withdrawals VALUES($1, $2, $3, TO_TIMESTAMP($4))"
		_, err := tx.ExecContext(
			request.Context(),
			addHistoryQuery, order_id, sum, name, time.Now().Unix(),
		)
		return err
	}
	err := general.RetryCode(f, syscall.ECONNREFUSED)
	if err != nil {
		return errors.New("Problem with assigning accrual " + err.Error())
	}
	return nil
}

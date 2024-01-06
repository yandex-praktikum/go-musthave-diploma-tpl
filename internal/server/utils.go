package server

import (
	"errors"
	"net/http"
	"syscall"

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

func ChangeAccrual(request *http.Request, reqData *WithdrawRequest, name string) error {
	f := func() error {
		changeBalance := "UPDATE balances SET current = current - $1, withdrawn = withdrawn + $1 WHERE name = $2"
		_, err := DB.ExecContext(
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

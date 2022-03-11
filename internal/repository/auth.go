package repository

import (
	"Loyalty/internal/models"
	"context"
	"fmt"
)

type auth struct {
	db DB
}

func NewAuth(db DB) *auth {
	return &auth{
		db: db,
	}
}

//save user in db ===========================================================
func (a *auth) SaveUser(user *models.User, accountNumber uint64) error {
	var number uint64
	q := `INSERT INTO users as u (login, password, account_id)
    VALUES($1,$2,(SELECT id FROM accounts WHERE number=$3))
	ON CONFLICT (login) DO UPDATE SET 
	login=EXCLUDED.login
   	RETURNING (SELECT number FROM accounts WHERE id=u.account_id)`
	a.db.QueryRow(context.Background(), q, user.Login, user.Password, accountNumber).Scan(&number)
	//internal db error
	if number == 0 {
		return ErrInt
	}
	//login already used
	if number != accountNumber {
		return fmt.Errorf(`%w`, ErrLoginConfl)
	}
	return nil
}

//get user from db ===========================================================
func (a *auth) GetUser(user *models.User) (uint64, error) {
	var number uint64
	q := `SELECT number FROM users
	JOIN 
		accounts ON users.account_id=accounts.id
	WHERE
		users.login=$1 AND users.password=$2;`
	a.db.QueryRow(context.Background(), q, user.Login, user.Password).Scan(&number)
	if number == 0 {
		return 0, ErrUsrUncor
	}
	return number, nil
}

package repository

import (
	"Loyalty/internal/models"
	"context"
	"errors"
	"fmt"
)

//save user in db
func (r *Repository) SaveUser(user *models.User, accountNumber string) error {
	var number string
	q := `INSERT INTO users as u (login,password,loyalty_account)
    VALUES($1,$2,(SELECT id FROM accounts WHERE number=$3))
	ON CONFLICT (login) DO UPDATE SET 
	login=EXCLUDED.login
   	RETURNING (SELECT number FROM accounts WHERE id=u.loyalty_account)`
	r.db.QueryRow(context.Background(), q, user.Login, user.Password, accountNumber).Scan(&number)
	//internal db error
	if number == "" {
		return errors.New("error: internal db error")
	}
	//login already used
	if number != accountNumber {
		err := errors.New("error: login conflict")
		return fmt.Errorf(`%w`, err)
	}
	return nil
}

//get user from db
func (r *Repository) GetUser(user *models.User) (string, error) {
	var number string
	q := `SELECT number FROM users
	JOIN 
		accounts ON users.loyalty_account=accounts.id
	WHERE
		users.login=$1 AND users.password=$2;`
	r.db.QueryRow(context.Background(), q, user.Login, user.Password).Scan(&number)
	if number == "" {
		return "", errors.New("error: username or password is not correct")
	}
	return number, nil
}

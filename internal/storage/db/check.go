package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
)

func (d *DateBase) CheckTableUserLogin(ctx context.Context, login string) error {
	var existingLogin string
	query := `SELECT login FROM users WHERE login = $1`

	err := d.storage.QueryRowContext(ctx, query, login).Scan(&existingLogin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return customerrors.ErrUserAlreadyExists
}

func (d *DateBase) CheckTableUserPassword(ctx context.Context, login string) (string, bool) {
	var existingPassword string
	query := `SELECT password FROM users WHERE login = $1`

	err := d.storage.QueryRowContext(ctx, query, login).Scan(&existingPassword)

	if err != nil {
		return "", false
	}

	if existingPassword == "" {
		return "", false
	}
	return existingPassword, true
}

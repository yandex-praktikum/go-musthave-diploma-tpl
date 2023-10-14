package auth

import (
	"database/sql"
	"errors"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/models"
)

func (c client) GetUserID(user models.User) (string, error) {
	var id string

	query := `SELECT id FROM users 
              WHERE login = $1
              AND password = $2`

	err := c.conn.QueryRow(query, user.Login, user.Password).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil

}

func (c client) SaveUser(user models.User) (string, error) {
	var id string

	query := `INSERT INTO users (login, password)
              VALUES ($1, $2)
              ON CONFLICT (login) DO NOTHING
              RETURNING id`

	err := c.conn.QueryRow(query, user.Login, user.Password).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("exists")
		}
		return "", err
	}

	return id, nil
}

func (c client) FindUserByID(id string) bool {
	query := `SELECT COUNT(*) FROM users WHERE id = $1`
	var count int

	row := c.conn.QueryRow(query, id)
	err := row.Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}

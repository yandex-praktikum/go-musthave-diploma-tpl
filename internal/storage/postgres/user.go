package postgres

import (
	"context"
	"database/sql"
	"github.com/A-Kuklin/gophermart/internal/domain/entities"
)

const (
	createUserSQL = `
		INSERT INTO gophermart.users (login, pass_hash)
		VALUES ($1, $2)
		RETURNING id, login, pass_hash
`
	getUserByLoginSQL = `
		SELECT id, login, pass_hash
		FROM gophermart.users
		WHERE login = $1
`
)

type AuthPSQL struct {
	db *sql.DB
}

func NewAuthPSQL(db *sql.DB) *AuthPSQL {
	return &AuthPSQL{
		db: db,
	}
}

func (s *AuthPSQL) CreateUser(ctx context.Context, user *entities.User) (*entities.User, error) {
	stmt, err := s.db.PrepareContext(ctx, createUserSQL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result := stmt.QueryRowContext(ctx, user.Login, user.PassHash)

	var usr entities.User
	err = result.Scan(&usr.ID, &usr.Login, &usr.PassHash)
	if err != nil {
		return nil, err
	}

	return &usr, nil
}

func (s *AuthPSQL) GetUserByLogin(ctx context.Context, user *entities.User) (*entities.User, error) {
	stmt, err := s.db.PrepareContext(ctx, getUserByLoginSQL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result := stmt.QueryRowContext(ctx, user.Login)

	var usr entities.User
	err = result.Scan(&usr.ID, &usr.Login, &usr.PassHash)
	if err != nil {
		return nil, err
	}

	return &usr, nil
}

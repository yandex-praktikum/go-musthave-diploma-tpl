package postgres

import (
	"context"

	"github.com/RedWood011/cmd/gophermart/internal/apperrors"
	"github.com/RedWood011/cmd/gophermart/internal/entity"
)

type User struct {
	ID       string `db:"id"`
	Login    string `db:"login"`
	Password string `db:"password"`
}

type UserBalance struct {
	Balance float32 `json:"current"`
	Spent   float32 `json:"withdrawn"`
}

func (r *Repository) GetUser(ctx context.Context, user entity.User) (entity.User, error) {
	var res User

	sqlCheckUser := `SELECT id FROM users WHERE login = $1;`
	query := r.db.QueryRow(ctx, sqlCheckUser, user.Login)
	err := query.Scan(&res.ID)
	if err != nil {
		return entity.User{}, apperrors.ErrUserNotFound
	}

	return entity.User{
		ID:       res.ID,
		Login:    res.Login,
		Password: res.Password,
	}, nil
}

func (r *Repository) GetUserBalance(ctx context.Context, userID string) (entity.UserBalance, error) {
	var result UserBalance
	sqlGetBalance := `SELECT balance, spend FROM users WHERE id = $1`
	query := r.db.QueryRow(ctx, sqlGetBalance, userID)
	err := query.Scan(&result.Balance, &result.Spent)
	if err != nil {
		return entity.UserBalance{}, err
	}
	return entity.UserBalance{
		Balance: result.Balance,
		Spent:   result.Spent,
	}, nil
}

func (r *Repository) SaveUser(ctx context.Context, user entity.User) error {
	sqlSaveUser := `INSERT INTO users (id,login, password) VALUES ($1,$2,$3)`
	_, err := r.db.Exec(ctx, sqlSaveUser, user.ID, user.Login, user.Password)
	return err
}

func (r *Repository) UpdateUsersBalance(ctx context.Context, users []entity.User) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	query := `UPDATE users SET balance = $2 WHERE id = $1`
	defer tx.Rollback(ctx)
	for _, value := range users {
		_, err = tx.Exec(ctx, query, value.ID, value.Balance)
		if err != nil {
			return err
		}
		tx.Commit(ctx)
	}
	return nil
}

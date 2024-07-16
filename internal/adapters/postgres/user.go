package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal/domain"
)

type PgUserRepository struct {
	db *pgxpool.Pool
}

var _ domain.UserRepository = (*PgUserRepository)(nil)

func NewPgUserRepository(db *pgxpool.Pool) *PgUserRepository {
	return &PgUserRepository{
		db: db,
	}
}

func (r *PgUserRepository) Insert(ctx context.Context, login, passwordHash string) (int, error) {
	var id int
	err := r.db.QueryRow(ctx, `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURN id`, login, passwordHash).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func (r *PgUserRepository) GetByID(ctx context.Context, id int) (domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx, `SELECT * FROM users WHERE id = $1`, id).Scan(&u.ID, &u.Login, &u.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, domain.ErrNotFound
	}
	return u, err
}

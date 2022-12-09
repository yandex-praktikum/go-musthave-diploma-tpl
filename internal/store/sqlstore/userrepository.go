package sqlstore

import (
	"database/sql"
	"github.com/iRootPro/gophermart/internal/entity"
	"github.com/iRootPro/gophermart/internal/store"
)

type UserRepository struct {
	store *Store
}

func (r *UserRepository) Create(u *entity.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	if err := u.BeforeCreate(); err != nil {
		return err
	}

	if err := r.store.db.QueryRow(""+
		"INSERT INTO users (username, encrypted_password) VALUES ($1, $2) RETURNING id",
		u.Username,
		u.EncryptedPassword).Scan(&u.ID); err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) FindByUsername(username string) (*entity.User, error) {
	u := &entity.User{}
	if err := r.store.db.QueryRow("SELECT id, username, encrypted_password FROM users WHERE username = $1",
		username).Scan(&u.ID, &u.Username, &u.EncryptedPassword); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	return u, nil
}

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
		"INSERT INTO users (login, encrypted_password) VALUES ($1, $2) RETURNING id",
		u.Login,
		u.EncryptedPassword).Scan(&u.ID); err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) FindByLogin(login string) (*entity.User, error) {
	u := &entity.User{}
	if err := r.store.db.QueryRow("SELECT id, login, encrypted_password FROM users WHERE login = $1",
		login).Scan(&u.ID, &u.Login, &u.EncryptedPassword); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	return u, nil
}

func (r *UserRepository) FindByID(id int) (*entity.User, error) {
	u := &entity.User{}
	if err := r.store.db.QueryRow("SELECT id, login, encrypted_password FROM users WHERE id = $1",
		id).Scan(&u.ID, &u.Login, &u.EncryptedPassword); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	return u, nil
}

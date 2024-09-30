package db

import "github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"

func (d *DateBase) Save(query string, args ...interface{}) error {
	_, err := d.storage.Exec(query, args...)
	if err != nil {
		return customerrors.ErrNotFound
	}

	return nil
}

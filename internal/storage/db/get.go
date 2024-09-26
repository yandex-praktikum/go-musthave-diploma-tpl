package db

import (
	"database/sql"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
)

func (d *DateBase) Get(query string, args ...interface{}) (*sql.Row, error) {
	row := d.storage.QueryRow(query, args...)
	if row == nil {
		return nil, customerrors.ErrNotFound
	}

	return row, nil
}

func (d *DateBase) Gets(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := d.storage.Query(query)
	if err != nil {
		return nil, err
	}

	return rows, err
}

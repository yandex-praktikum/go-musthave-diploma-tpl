package postgres

import (
	"database/sql"
)

// scanRows — универсальная функция для сканирования rows в слайс элементов типа T.
func scanRows[T any](rows *sql.Rows, scanFunc func(*sql.Rows) (T, error)) ([]T, error) {
	defer rows.Close()
	var results []T
	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

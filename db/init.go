package db

import (
	"context"
	"database/sql"
)

func InitDB(ctx context.Context, db *sql.DB) error {
	for _, query := range table {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

package gophermart

import (
	"context"
	"database/sql"
	"os"
)

type (
	PgMigrator struct {
		pg *sql.DB
	}

	MigratorInterface interface {
		Apply(ctx context.Context, dsn string, files []string) error
	}
)

func NewPgMigrator(psql *sql.DB) MigratorInterface {
	return &PgMigrator{pg: psql}
}

func (m *PgMigrator) Apply(ctx context.Context, dir string, files []string) error {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			// if not exists, it's not an error
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		_, err = m.pg.ExecContext(ctx, string(content))
		if err != nil {
			return err
		}
	}

	return nil
}

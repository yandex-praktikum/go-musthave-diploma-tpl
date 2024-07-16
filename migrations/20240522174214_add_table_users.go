package migrations

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddTableUsers, downAddTableUsers)
}

func upAddTableUsers(ctx context.Context, tx *sql.Tx) error {
	// TODO
	return nil
}

func downAddTableUsers(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}

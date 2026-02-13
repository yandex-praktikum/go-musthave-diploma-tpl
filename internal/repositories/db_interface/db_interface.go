package dbinterface

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type DbIface interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

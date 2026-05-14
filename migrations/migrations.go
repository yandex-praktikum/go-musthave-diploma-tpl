package migrations

import (
	"context"
	"embed"
	"fmt"

	"github.com/AnxVit/go-musthave-diploma-tpl/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

//go:embed pgmigrations/*.sql
var migrationFiles embed.FS

func Migrate(dbstring, command string, arguments []string) {
	db, err := goose.OpenDBWithDriver("postgres", dbstring)
	if err != nil {
		logger.Log.Warn("goose: failed to open DB", zap.Error(err))
		return
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Log.Warn("goose: failed to close DB:", zap.Error(err))
		}
	}()

	goose.SetBaseFS(migrationFiles)

	ctx := context.Background()
	if err := goose.RunContext(ctx, command, db, "pgmigrations", arguments...); err != nil {
		logger.Log.Warn(fmt.Sprintf("goose %v", command), zap.Error(err))
	}
}

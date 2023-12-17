package storage

import (
	"database/sql"

	"github.com/benderr/gophermart/internal/config"
	"github.com/benderr/gophermart/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func MustLoad(config *config.Config, logger logger.Logger) *sql.DB {
	if len(config.DatabaseDsn) == 0 {
		logger.Errorln("[DB]: database dsn not specified")
		panic("database dsn not specified")
	}

	db, dberr := sql.Open("pgx", config.DatabaseDsn)
	if dberr != nil {
		logger.Errorln("[DB]: database dsn not specified", dberr)
		db.Close()
		panic(dberr)
	}

	runMigration()

	return db
}

func runMigration() {

}

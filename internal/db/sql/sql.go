package sql

import (
	"context"
	"errors"
	"fmt"
	"github.com/GTech1256/go-musthave-diploma-tpl/pkg/retry"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQL struct {
	DB *pgxpool.Pool
}

var (
	ErrConfig     = errors.New("неправильный путь для подключения")
	ErrCreatePool = errors.New("пул соединений не получилось создать")
	ErrPing       = errors.New("ping не прошел")
)

func NewSQL(host string) (*SQL, error) {
	config, err := pgxpool.ParseConfig(host)
	if err != nil {
		return nil, ErrConfig
	}

	ctx := context.Background()
	var db *pgxpool.Pool

	err = retry.MakeRetry(func() error {
		db, err = pgxpool.NewWithConfig(ctx, config)

		return err
	})
	if err != nil {
		return nil, ErrCreatePool
	}

	if err = pingWithRetry(ctx, db); err != nil {
		return nil, ErrPing
	}

	s := &SQL{
		DB: db,
	}

	err = s.MigrateDown(host)
	if err != nil {
		fmt.Println("ERROR: MigrationDown", err)
	}

	err = s.MigrateUp(host)
	if err != nil {
		fmt.Println("ERROR: MigrationUp", err)
	}

	return s, nil
}

func pingWithRetry(ctx context.Context, db *pgxpool.Pool) error {
	return retry.MakeRetry(func() error {
		return db.Ping(ctx)
	})
}

func (q SQL) MigrateUp(dataSourceName string) error {
	fmt.Println("Migration Up Started")
	m, err := migrate.New(
		"file://internal/config/db/migrations",
		dataSourceName)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil {
		return err
	}

	fmt.Println("Migration Up Ended")
	return nil
}

func (q SQL) MigrateDown(dataSourceName string) error {
	fmt.Println("Migration Down Started")
	m, err := migrate.New(
		"file://internal/config/db/migrations",
		dataSourceName)
	if err != nil {
		return err
	}
	err = m.Down()
	if err != nil {
		return err
	}

	fmt.Println("Migration Down Ended")
	return nil
}

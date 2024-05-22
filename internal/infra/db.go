package infra

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal"
	_ "github.com/sashaaro/go-musthave-diploma-tpl/migrations"

	"log"
)

func CreatePgxPool() *pgxpool.Pool {
	config, err := pgxpool.ParseConfig(internal.Config.DatabaseDSN)
	if err != nil {
		log.Fatalf("can't parse config: %v", err)
	}

	if err != nil {
		log.Fatal("can't connect to database", err)
	}

	db := stdlib.OpenDB(*config.ConnConfig)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal("can't set dialect: ", err)
	}

	if err := goose.Up(db, "./"); err != nil {
		log.Fatal("can't run migrations: ", err)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal("can't create pool: ", err)
	}
	return pool
}

package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/AnxVit/go-musthave-diploma-tpl/internal/logger"
)

type PostgresClient struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, url string) *PostgresClient {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		logger.Log.Warn("Couldn't connect to database", zap.Error(err))
		return nil
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		logger.Log.Warn("Couldn't connect to database", zap.Error(err))
		pool.Close()
	} else {
		conn.Release()
	}

	return &PostgresClient{
		pool: pool,
	}
}

func (p *PostgresClient) Close() {
	if p.pool != nil {
		p.pool.Close()
		logger.Log.Info("Database connection pool closed")
	}
}

func (p *PostgresClient) Pool() *pgxpool.Pool {
	return p.pool
}

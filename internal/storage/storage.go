package storage

import (
	"context"
	"database/sql"
	"sync"

	"github.com/brisk84/gofemart/domain"
	"github.com/brisk84/gofemart/internal/config"
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
	"go.uber.org/zap"
)

type storage struct {
	logger *zap.Logger

	users map[string]domain.User

	orders map[int64]domain.Order

	currentUserID  int64
	currentOrderID int64

	userMtx  sync.RWMutex
	orderMtx sync.RWMutex

	db      *sql.DB
	mainDsn string
}

func New(logger *zap.Logger, cfg config.Config) *storage {
	return &storage{
		logger: logger,
		users:  make(map[string]domain.User),

		orders:  make(map[int64]domain.Order),
		mainDsn: cfg.DatabaseDSN,
	}
}

func (s *storage) Connect(ctx context.Context) error {
	var err error
	s.db, err = sql.Open("postgres", s.mainDsn)
	if err != nil {
		return err
	}
	err = s.db.PingContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) MigrateUp(ctx context.Context) error {
	return goose.Up(s.db, "internal/storage/migrations")
}

func (s *storage) MigrateDown(ctx context.Context) error {
	return goose.Down(s.db, "internal/storage/migrations")
}

func (s *storage) Close() error {
	return s.db.Close()
}

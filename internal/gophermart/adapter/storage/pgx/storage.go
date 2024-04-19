package pgx

import (
	"context"
	"sync"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/config"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
)

func NewStorage(appCnf context.Context, logger domain.Logger, gConf *config.GophermartConfig) *storage {
	once.Do(func() {
		st = initializePGXConf(appCnf, logger, gConf)
	})

	return st
}

var once sync.Once
var st *storage

func initializePGXConf(ctx context.Context, logger domain.Logger, gConf *config.GophermartConfig) *storage {
	logger.Infow("initializePGXConf", "status", "start")

	pConf, err := pgxpool.ParseConfig(gConf.DatabaseUri)
	if err != nil {
		panic(err)
	}

	// Конфигурация по мотивам
	// https://habr.com/ru/companies/oleg-bunin/articles/461935/
	pConf.MaxConns = int32(gConf.MaxConns)
	pConf.ConnConfig.RuntimeParams["standard_conforming_strings"] = "on"
	pConf.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pConf.MaxConnLifetime = 5 * time.Minute
	pConf.MaxConnIdleTime = 5 * time.Minute

	pConf.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger: &loggerAdapter{
			logger: logger,
		},
		LogLevel: tracelog.LogLevelTrace,
	}

	pPool, err := pgxpool.NewWithConfig(ctx, pConf)

	if err != nil {
		panic(err)
	}

	st = &storage{
		pPool:                pPool,
		logger:               logger,
		processingLimit:      gConf.ProcessingLimit,
		processingScoreDelta: gConf.ProcessingScoreDelta,
	}

	st.init(ctx)

	st.logger.Infow("initializePGXConf", "status", "complete")
	return st
}

type storage struct {
	pPool                *pgxpool.Pool
	logger               domain.Logger
	processingLimit      int
	processingScoreDelta time.Duration
}

func (st *storage) init(ctx context.Context) error {
	st.logger.Infow("init", "status", "start")

	tx, err := st.pPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// с индексами не разбирался TODO
	tx.Exec(ctx, `
	create table if not exists userInfo(
		userId serial,
		login text not null,
		hash text not null,
		salt text not null,
		primary key(userId),
		unique (login)
	);`)

	tx.Exec(ctx, `
	create table if not exists orderData(
		number varchar(255),
		userId int not null,
		status varchar(255),
		accrual float8,
		uploaded_at timestamptz,
		score timestamptz not null default now(),
		primary key(number),
		foreign key (userId) references userInfo(userId)
	);`)

	tx.Exec(ctx, `
	create table if not exists balance(
		balanceId serial,
		userId int not null,
		current float8 not null default 0,
		withdrawn float8 not null default 0,
		release int not null default 0,
		primary key(balanceId),
		unique (userId),
		foreign key (userId) references userInfo(userId)
		);`)

	tx.Exec(ctx, `
	create table if not exists withdrawal(
		withdrawalId serial,
		balanceId int not null,
		number varchar(255) not null,
		sum float8 not null,
		processed_at  timestamptz not null default now(),
		primary key(withdrawalId),
		foreign key (balanceId) references balance(balanceId)
		);`)

	return tx.Commit(ctx)
}

func (st *storage) Ping(ctx context.Context) error {
	return st.pPool.Ping(ctx)
}

package pgx

import (
	"context"
	"errors"
	"sync"

	"github.com/StasMerzlyakov/gophermart/internal/config"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		pPool:  pPool,
		logger: logger,
	}

	st.init(ctx)

	st.logger.Infow("initializePGXConf", "status", "complete")
	return st
}

type storage struct {
	pPool  *pgxpool.Pool
	logger domain.Logger
}

func (st *storage) init(ctx context.Context) error {
	st.logger.Infow("init", "status", "start")

	tx, err := st.pPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tx.Exec(ctx, `
	create table if not exists userInfo(
		userId serial,
		login text not null,
		hash text not null,
		salt text not null,
		primary key(userId),
		unique (login)
	);	
	`)

	return tx.Commit(ctx)

	/*if db, err := sql.Open("pgx", st.databaseURL); err != nil {
		st.logger.Infow("Bootstrap", "status", "error", "msg", err.Error())
		return err
	} else {
		st.db = db
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS counter(
			name text not null,
			value bigint,
			PRIMARY KEY(name)
		);`)

		tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS gauge(
			name text not null,
			value double precision,
			PRIMARY KEY(name)
		);`)

		return tx.Commit()
	} */

}

func (st *storage) Ping(ctx context.Context) error {
	return st.pPool.Ping(ctx)
}

func (st *storage) RegisterUser(ctx context.Context, ld *domain.LoginData) error {
	st.logger.Infow("pgx.RegisterUser", "status", "start")

	var userID int
	if err := st.pPool.QueryRow(ctx,
		"insert into userInfo(login, hash, salt) values ($1, $2, $3) returning userId",
		ld.Login,
		ld.Hash,
		ld.Salt).Scan(&userID); err == nil {
		st.logger.Infow("pgx.RegisterUser", "status", "success", "userID", userID)
		return nil
	} else {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return domain.ErrLoginIsBusy
			}
		}
		st.logger.Errorw("pgx.Register", "err", err.Error())
		return domain.ErrServerInternal
	}
}

func (st *storage) GetUserData(ctx context.Context, login string) (*domain.LoginData, error) {
	st.logger.Infow("pgx.GetUserData", "status", "start")

	var data domain.LoginData
	err := st.pPool.QueryRow(ctx, "select userId, login, hash, salt from userInfo where login = $1", login).Scan(&data.UserID, &data.Login, &data.Hash, &data.Salt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			st.logger.Infow("pgx.GetUserData", "status", "not found", "login", login)
			return nil, nil
		}
		st.logger.Errorw("pgx.Register", "err", err.Error())
		return nil, domain.ErrServerInternal
	}

	return &data, nil
}

package pgx

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

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

	tx.Exec(ctx, `
	create table if not exists userInfo(
		userId serial,
		login text not null,
		hash text not null,
		salt text not null,
		primary key(userId),
		unique (login)
	);`)

	// TODO индексы по status, userId
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

	return tx.Commit(ctx)
}

func (st *storage) Ping(ctx context.Context) error {
	return st.pPool.Ping(ctx)
}

func (st *storage) RegisterUser(ctx context.Context, ld *domain.LoginData) (int, error) {
	st.logger.Infow("pgx.RegisterUser", "status", "start")

	var userID int
	if err := st.pPool.QueryRow(ctx,
		"insert into userInfo(login, hash, salt) values ($1, $2, $3) returning userId",
		ld.Login,
		ld.Hash,
		ld.Salt).Scan(&userID); err == nil {
		st.logger.Infow("pgx.RegisterUser", "status", "success", "userID", userID)
		return userID, nil
	} else {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return -1, domain.ErrLoginIsBusy
			}
		}
		st.logger.Errorw("pgx.Register", "err", err.Error())
		return -1, domain.ErrServerInternal
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

func (st *storage) Upload(ctx context.Context, data *domain.OrderData) error {

	if data == nil {
		st.logger.Errorw("pgx.Upload", "err", "data is nil")
		return domain.ErrServerInternal
	}

	var number domain.OrderNumber

	if err := st.pPool.QueryRow(ctx,
		`insert into orderData(number, userId, status, uploaded_at) values ($1, $2, $3, $4) 
		on conflict("number") do nothing returning number;
	  `, data.Number, data.UserID, data.Status, time.Time(data.UploadedAt).UTC()).Scan(&number); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// запись с таким number уже есть; проверим какому пользователю принадлежит
			var userId int
			err = st.pPool.QueryRow(ctx, `select userId from orderData where number = $1`, data.Number).Scan(&userId)
			if err != nil {
				st.logger.Infow("pgx.Upload", "err", err.Error())
				return domain.ErrServerInternal
			}
			if userId == data.UserID {
				return domain.ErrOrderNumberAlreadyUploaded
			} else {
				return domain.ErrDublicateOrderNumber
			}
		} else {
			st.logger.Infow("pgx.Upload", "err", err.Error())
			return domain.ErrServerInternal
		}
	}

	return nil
}

func (st *storage) Orders(ctx context.Context, userID int) ([]domain.OrderData, error) {
	var orders []domain.OrderData

	rows, err := st.pPool.Query(ctx,
		`select number, userId, status, accrual, uploaded_at from orderData where userId = $1`,
		userID,
	)

	if err != nil {
		st.logger.Infow("pgx.Orders", "err", err.Error())
		return nil, domain.ErrServerInternal
	}

	defer rows.Close()

	for rows.Next() {
		var data domain.OrderData
		var uploaded time.Time
		err = rows.Scan(&data.Number, &data.UserID, &data.Status, &data.Accrual, &uploaded)
		if err != nil {
			st.logger.Infow("pgx.ForProcessing", "err", err.Error())
			return nil, domain.ErrServerInternal
		}
		data.UploadedAt = domain.RFC3339Time(uploaded)
		orders = append(orders, data)
	}

	err = rows.Err()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			st.logger.Infow("pgx.Orders", "status", "not found")
			return nil, domain.ErrNotFound
		}
		st.logger.Infow("pgx.Orders", "err", err.Error())
		return nil, domain.ErrServerInternal
	}

	return orders, nil
}

func (st *storage) Update(ctx context.Context, number domain.OrderNumber, status domain.OrderStatus, accrual *float64) error {
	rows, err := st.pPool.Query(ctx,
		`update orderData set status = $1, accrual = $2 where number = $3`,
		string(status), accrual, string(number),
	)

	if err != nil {
		st.logger.Infow("pgx.Update", "err", err.Error())
		return domain.ErrServerInternal
	}

	defer rows.Close()

	if rows.Next() {
		return nil
	}

	err = rows.Err()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			st.logger.Infow("pgx.Update", "status", "not found")
			return domain.ErrNotFound
		}
		return fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
	}
	return nil
}

func (st *storage) ForProcessing(ctx context.Context, statuses []domain.OrderStatus) ([]domain.OrderData, error) {

	var forProcessing []domain.OrderData

	var sStatus []string
	for _, s := range statuses {
		sStatus = append(sStatus, string(s))
	}

	rows, err := st.pPool.Query(ctx,
		`update orderData set score = $1 
		 where number in 
		   (select number from orderdata where status = ANY($2) and score < $3 limit $4) 
		 returning 
		    number, userId, status, accrual, uploaded_at;`,
		time.Now().Add(st.processingScoreDelta),
		sStatus,
		time.Now(),
		st.processingLimit,
	)

	if err != nil {
		st.logger.Infow("pgx.ForProcessing", "err", err.Error())
		return nil, domain.ErrServerInternal
	}

	defer rows.Close()

	for rows.Next() {
		var data domain.OrderData
		var uploaded time.Time
		err = rows.Scan(&data.Number, &data.UserID, &data.Status, &data.Accrual, &uploaded)
		if err != nil {
			st.logger.Infow("pgx.ForProcessing", "err", err.Error())
			return nil, domain.ErrServerInternal
		}
		data.UploadedAt = domain.RFC3339Time(uploaded)
		forProcessing = append(forProcessing, data)
	}

	err = rows.Err()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			st.logger.Infow("pgx.ForProcessing", "status", "not found")
			return nil, nil
		}
		st.logger.Infow("pgx.ForProcessing", "err", err.Error())
		return nil, domain.ErrServerInternal
	}
	st.logger.Infow("pgx.ForProcessing", "status", "found", "count", len(forProcessing))

	return forProcessing, nil
}

func (st *storage) UpdateBatch(ctx context.Context, orders []domain.OrderData) error {

	tx, err := st.pPool.Begin(ctx)
	if err != nil {
		st.logger.Errorw("storage.UpdateBatch", "err", err.Error())
		return domain.ErrServerInternal
	}

	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	for _, ord := range orders {
		batch.QueuedQueries = append(batch.QueuedQueries,
			&pgx.QueuedQuery{
				SQL:       `update orderData set status = $1, accrual = $2 where number = $3`,
				Arguments: []any{string(ord.Status), ord.Accrual, string(ord.Number)},
			},
		)
	}

	err = tx.SendBatch(context.Background(), batch).Close()

	if err != nil {
		st.logger.Infow("storage.UpdateBatch", "err", err.Error())
		return domain.ErrServerInternal
	}

	return nil
}

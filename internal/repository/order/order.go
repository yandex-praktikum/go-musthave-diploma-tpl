package sql

import (
	"context"
	"errors"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type storage struct {
	logger logging.Logger
	db     DB
}

type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Ping(ctx context.Context) error
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

var (
	ErrNotUniqueLogin                  = errors.New("пользователь с таким логином уже зарегистрирован")
	ErrInvalidLoginPasswordCombination = errors.New("неверная пара логин/пароль")
)

func NewStorage(db DB, logger logging.Logger) *storage {
	return &storage{
		db:     db,
		logger: logger,
	}
}

func (s *storage) Create(ctx context.Context, userId int, orderNumber *entity.OrderNumber) (*entity.OrderDB, error) {
	var orderDB entity.OrderDB
	err := s.db.QueryRow(
		ctx,
		"INSERT INTO gophermart.orders (number, status, user_id) values ($1, $2, $3) RETURNING id, number, status, accrual, uploaded_at, user_id",
		*orderNumber,
		"REGISTERED",
		userId,
	).Scan(&orderDB.ID, &orderDB.Number, &orderDB.Status, &orderDB.Accrual, &orderDB.UploadedAt, &orderDB.UserId)

	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	return &orderDB, nil
}

func (s *storage) GetByOrderId(ctx context.Context, orderId int) (*entity.OrderDB, error) {
	var orderDB entity.OrderDB

	err := s.db.QueryRow(
		ctx,
		"SELECT id, number, status, accrual, uploaded_at, user_id FROM gophermart.orders WHERE id = $1",
		orderId,
	).Scan(&orderDB.ID, &orderDB.Number, &orderDB.Status, &orderDB.Accrual, &orderDB.UploadedAt, &orderDB.UserId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	return &orderDB, nil
}

func (s *storage) GetByOrderNumber(ctx context.Context, orderNumber *entity.OrderNumber) (*entity.OrderDB, error) {
	var orderDB entity.OrderDB

	err := s.db.QueryRow(
		ctx,
		"SELECT id, number, status, accrual, uploaded_at, user_id FROM gophermart.orders WHERE number = $1",
		*orderNumber,
	).Scan(&orderDB.ID, &orderDB.Number, &orderDB.Status, &orderDB.Accrual, &orderDB.UploadedAt, &orderDB.UserId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	return &orderDB, nil
}

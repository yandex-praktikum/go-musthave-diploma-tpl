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
	ErrNoRowsForUpdate = errors.New("нет записей для обновления")
)

func NewStorage(db DB, logger logging.Logger) *storage {
	return &storage{
		db:     db,
		logger: logger,
	}
}

func (s *storage) Create(ctx context.Context, userID int, orderNumber *entity.OrderNumber) (*entity.OrderDB, error) {
	var orderDB entity.OrderDB

	err := s.db.QueryRow(
		ctx,
		"INSERT INTO gophermart.orders (number, status, user_id) values ($1, $2, $3) RETURNING id, number, status, accrual, uploaded_at, user_id",
		*orderNumber,
		"REGISTERED",
		userID,
	).Scan(&orderDB.ID, &orderDB.Number, &orderDB.Status, &orderDB.Accrual, &orderDB.UploadedAt, &orderDB.UserID)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	return &orderDB, nil
}

func (s *storage) GetByOrderID(ctx context.Context, orderID int) (*entity.OrderDB, error) {
	var orderDB entity.OrderDB

	err := s.db.QueryRow(
		ctx,
		"SELECT id, number, status, accrual, uploaded_at, user_id FROM gophermart.orders WHERE id = $1",
		orderID,
	).Scan(&orderDB.ID, &orderDB.Number, &orderDB.Status, &orderDB.Accrual, &orderDB.UploadedAt, &orderDB.UserID)
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
	).Scan(&orderDB.ID, &orderDB.Number, &orderDB.Status, &orderDB.Accrual, &orderDB.UploadedAt, &orderDB.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	return &orderDB, nil
}

func (s *storage) GetOrdersForProcessing(ctx context.Context) ([]*entity.OrderDB, error) {
	orderDBs := make([]*entity.OrderDB, 0)

	row, err := s.db.Query(
		ctx,
		"SELECT id, number, status, accrual, uploaded_at, user_id FROM gophermart.orders WHERE status IN ('REGISTERED', 'PROCESSING')",
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return orderDBs, nil
	}
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	for row.Next() {
		var orderDB entity.OrderDB
		err := row.Scan(&orderDB.ID, &orderDB.Number, &orderDB.Status, &orderDB.Accrual, &orderDB.UploadedAt, &orderDB.UserID)
		if err != nil {
			s.logger.Error(err)
			continue
		}

		orderDBs = append(orderDBs, &orderDB)
	}

	return orderDBs, nil
}

func (s *storage) Update(ctx context.Context, orderDB *entity.OrderDB) error {
	_, err := s.db.Exec(
		ctx,
		"UPDATE gophermart.orders SET status=$2, accrual=$3 WHERE id = $1",
		orderDB.ID,
		orderDB.Status,
		orderDB.Accrual,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoRowsForUpdate
	}
	if err != nil {
		s.logger.Error(err)
		return err
	}

	return nil
}

func (s *storage) GetOrdersByUserID(ctx context.Context, userID int) ([]*entity.OrderDB, error) {
	orderDBs := make([]*entity.OrderDB, 0)

	row, err := s.db.Query(
		ctx,
		"SELECT id, number, status, accrual, uploaded_at, user_id FROM gophermart.orders WHERE user_id = $1",
		userID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return orderDBs, nil
	}
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	for row.Next() {
		var orderDB entity.OrderDB
		err := row.Scan(&orderDB.ID, &orderDB.Number, &orderDB.Status, &orderDB.Accrual, &orderDB.UploadedAt, &orderDB.UserID)
		if err != nil {
			s.logger.Error(err)
			continue
		}

		orderDBs = append(orderDBs, &orderDB)
	}

	return orderDBs, nil
}

package repository

import (
	"Loyalty/internal/models"
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//this interface implements pgx.Conn, pgx.Pool and pgx.Mock
type DB interface {
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Ping(context.Context) error
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
}

func NewDB(ctx context.Context, address string) (*pgxpool.Pool, error) {
	//init connection
	conn, err := pgxpool.Connect(ctx, address)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func AutoMigration(isAllowed bool, address string) error {

	if !isAllowed {
		return nil
	}
	//open connection
	db, err := gorm.Open(postgres.Open(address), &gorm.Config{})
	if err != nil {
		return err
	}

	//run automigration
	if err := db.AutoMigrate(&models.User{}, &models.Account{}, &models.Order{}, &models.Withdrawal{}); err != nil {
		return err
	}
	db.Exec("ALTER TABLE users ADD CONSTRAINT loyalty_account_fk FOREIGN KEY (loyalty_account) REFERENCES accounts(id)")
	db.Exec("ALTER TABLE orders ADD CONSTRAINT user_id_fk FOREIGN KEY (user_id) REFERENCES users(id)")
	db.Exec("ALTER TABLE withdrawals ADD CONSTRAINT order_id_fk FOREIGN KEY (order_id) REFERENCES orders(id)")

	return nil
}

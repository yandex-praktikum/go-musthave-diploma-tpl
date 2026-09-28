package repository

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/true18/go-musthave-diploma-tpl/internal/migrations"
	"github.com/true18/go-musthave-diploma-tpl/internal/money"
)

func TestCreateUser(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectQuery(createUserQuery).
		WithArgs("alice", "hash").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	id, err := New(db).CreateUser(context.Background(), "alice", []byte("hash"))
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Fatalf("id = %d", id)
	}
	assertExpectations(t, mock)
}

func TestCreateUserDuplicate(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectQuery(createUserQuery).
		WithArgs("alice", "hash").
		WillReturnError(&pgconn.PgError{Code: pgUniqueViolation})

	_, err := New(db).CreateUser(context.Background(), "alice", []byte("hash"))
	if !errors.Is(err, ErrUserExists) {
		t.Fatalf("err = %v", err)
	}
	assertExpectations(t, mock)
}

func TestFindUserByLogin(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	createdAt := time.Now()
	mock.ExpectQuery(findUserQuery).
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password_hash", "created_at"}).
			AddRow(1, "alice", "hash", createdAt))

	user, err := New(db).FindUserByLogin(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != 1 || user.Login != "alice" || string(user.PasswordHash) != "hash" {
		t.Fatalf("unexpected user: %+v", user)
	}
	assertExpectations(t, mock)
}

func TestCreateOrder(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectQuery(findOrderOwnerQuery).
		WithArgs("12345678903").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(createOrderQuery).
		WithArgs("12345678903", int64(1), StatusNew).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := New(db).CreateOrder(context.Background(), 1, "12345678903")
	if err != nil {
		t.Fatal(err)
	}
	assertExpectations(t, mock)
}

func TestCreateOrderDuplicateSameUser(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectQuery(findOrderOwnerQuery).
		WithArgs("12345678903").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

	err := New(db).CreateOrder(context.Background(), 1, "12345678903")
	if !errors.Is(err, ErrOrderExistsSameUser) {
		t.Fatalf("err = %v", err)
	}
	assertExpectations(t, mock)
}

func TestCreateOrderDuplicateAnotherUser(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectQuery(findOrderOwnerQuery).
		WithArgs("12345678903").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(2))

	err := New(db).CreateOrder(context.Background(), 1, "12345678903")
	if !errors.Is(err, ErrOrderExistsAnotherUser) {
		t.Fatalf("err = %v", err)
	}
	assertExpectations(t, mock)
}

func TestListOrders(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(listOrdersQuery).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"}).
			AddRow("2", StatusProcessed, "500.00", now).
			AddRow("1", StatusProcessing, nil, now.Add(-time.Hour)))

	orders, err := New(db).ListUserOrders(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) != 2 || orders[0].Accrual == nil || *orders[0].Accrual != mustAmount(t, "500") {
		t.Fatalf("unexpected orders: %+v", orders)
	}
	assertExpectations(t, mock)
}

func TestUpdateOrderStatusAndAccrual(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	amount := mustAmount(t, "500")
	mock.ExpectExec(updateOrderQuery).
		WithArgs("12345678903", StatusProcessed, "500", StatusInvalid, StatusProcessed).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := New(db).UpdateOrder(context.Background(), "12345678903", StatusProcessed, &amount); err != nil {
		t.Fatal(err)
	}
	assertExpectations(t, mock)
}

func TestBalanceCalculation(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectQuery(balanceQuery).
		WithArgs(int64(1), StatusProcessed).
		WillReturnRows(sqlmock.NewRows([]string{"accrued", "withdrawn"}).AddRow("600.50", "100.50"))

	balance, err := New(db).Balance(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if balance.Current != mustAmount(t, "500") || balance.Withdrawn != mustAmount(t, "100.5") {
		t.Fatalf("unexpected balance: %+v", balance)
	}
	assertExpectations(t, mock)
}

func TestWithdraw(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(lockUserQuery).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(balanceQuery).WithArgs(int64(1), StatusProcessed).
		WillReturnRows(sqlmock.NewRows([]string{"accrued", "withdrawn"}).AddRow("1000.00", "0"))
	mock.ExpectExec(insertWithdrawalQuery).WithArgs(int64(1), "2377225624", "751").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := New(db).Withdraw(context.Background(), 1, "2377225624", mustAmount(t, "751"))
	if err != nil {
		t.Fatal(err)
	}
	assertExpectations(t, mock)
}

func TestWithdrawInsufficientFunds(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(lockUserQuery).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(balanceQuery).WithArgs(int64(1), StatusProcessed).
		WillReturnRows(sqlmock.NewRows([]string{"accrued", "withdrawn"}).AddRow("10.00", "0"))
	mock.ExpectRollback()

	err := New(db).Withdraw(context.Background(), 1, "2377225624", mustAmount(t, "751"))
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("err = %v", err)
	}
	assertExpectations(t, mock)
}

func TestListWithdrawals(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(listWithdrawalsQuery).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"order_number", "sum", "processed_at"}).
			AddRow("2377225624", "500.00", now))

	withdrawals, err := New(db).ListWithdrawals(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(withdrawals) != 1 || withdrawals[0].Sum != mustAmount(t, "500") {
		t.Fatalf("unexpected withdrawals: %+v", withdrawals)
	}
	assertExpectations(t, mock)
}

func TestPostgresIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URI")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URI is not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := migrations.Up(db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`TRUNCATE withdrawals, orders, users RESTART IDENTITY`); err != nil {
		t.Fatal(err)
	}

	repo := New(db)
	userID, err := repo.CreateUser(context.Background(), "alice", []byte("hash"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindUserByLogin(context.Background(), "alice"); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateOrder(context.Background(), userID, "12345678903"); err != nil {
		t.Fatal(err)
	}
	amount := mustAmount(t, "500")
	if err := repo.UpdateOrder(context.Background(), "12345678903", StatusProcessed, &amount); err != nil {
		t.Fatal(err)
	}
	if err := repo.Withdraw(context.Background(), userID, "2377225624", mustAmount(t, "100")); err != nil {
		t.Fatal(err)
	}
	balance, err := repo.Balance(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if balance.Current != mustAmount(t, "400") || balance.Withdrawn != mustAmount(t, "100") {
		t.Fatalf("unexpected balance: %+v", balance)
	}
}

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	return db, mock
}

func assertExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func mustAmount(t *testing.T, value string) money.Amount {
	t.Helper()

	amount, err := money.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	return amount
}

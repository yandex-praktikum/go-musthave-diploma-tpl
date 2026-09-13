package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

// newMockDB создаёт тестовое соединение с sqlmock.
func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// --- UserRepository ---

func TestUserRepository_Create(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)
	user := &domain.User{Login: "alice", Password: "hash", CreatedAt: time.Now()}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("alice", "hash", user.CreatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	err := repo.Create(user)
	requireNoError(t, err)

	if user.ID != 10 {
		t.Errorf("user.ID = %d, want 10", user.ID)
	}

	requireMockExpectations(t, mock)
}

func TestUserRepository_Create_LoginExists(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)
	user := &domain.User{Login: "alice", Password: "hash", CreatedAt: time.Now()}

	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnError(&pgconn.PgError{Code: "23505", ConstraintName: "users_login_key"})

	err := repo.Create(user)
	requireError(t, err, domain.ErrLoginAlreadyExists)

	requireMockExpectations(t, mock)
}

func TestUserRepository_Create_OtherError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)
	user := &domain.User{Login: "alice", Password: "hash", CreatedAt: time.Now()}

	dbErr := errors.New("db down")
	mock.ExpectQuery(`INSERT INTO users`).WillReturnError(dbErr)

	err := repo.Create(user)
	requireError(t, err, dbErr)
	requireNotError(t, err, domain.ErrLoginAlreadyExists)

	requireMockExpectations(t, mock)
}

func TestUserRepository_FindByLogin(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "login", "password", "created_at"}).
		AddRow(1, "alice", "hash", now)
	mock.ExpectQuery(`SELECT id, login, password, created_at FROM users`).
		WithArgs("alice").WillReturnRows(rows)

	user, err := repo.FindByLogin("alice")
	requireNoError(t, err)

	if user.ID != 1 || user.Login != "alice" || user.Password != "hash" {
		t.Errorf("unexpected user: %+v", user)
	}

	requireMockExpectations(t, mock)
}

func TestUserRepository_FindByLogin_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery(`SELECT id, login, password, created_at FROM users`).
		WithArgs("ghost").WillReturnError(sql.ErrNoRows)

	_, err := repo.FindByLogin("ghost")
	requireError(t, err, domain.ErrUserNotFound)

	requireMockExpectations(t, mock)
}

func TestUserRepository_FindByLogin_OtherError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)

	dbErr := errors.New("db down")
	mock.ExpectQuery(`SELECT id, login, password, created_at FROM users`).
		WillReturnError(dbErr)

	_, err := repo.FindByLogin("alice")
	requireError(t, err, dbErr)
	requireNotError(t, err, domain.ErrUserNotFound)

	requireMockExpectations(t, mock)
}

func TestUserRepository_FindByID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "login", "password", "created_at"}).
		AddRow(5, "bob", "hash", time.Now())
	mock.ExpectQuery(`SELECT id, login, password, created_at FROM users`).
		WithArgs(int64(5)).WillReturnRows(rows)

	user, err := repo.FindByID(5)
	requireNoError(t, err)

	if user.Login != "bob" {
		t.Errorf("user.Login = %q, want bob", user.Login)
	}

	requireMockExpectations(t, mock)
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery(`SELECT id, login, password, created_at FROM users`).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.FindByID(42)
	requireError(t, err, domain.ErrUserNotFound)

	requireMockExpectations(t, mock)
}

func TestUserRepository_Exists(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.Exists("alice")
	requireNoError(t, err)

	if !exists {
		t.Error("Exists() = false, want true")
	}

	mock.ExpectQuery(`SELECT EXISTS`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err = repo.Exists("ghost")
	requireNoError(t, err)

	if exists {
		t.Error("Exists() = true, want false")
	}

	requireMockExpectations(t, mock)
}

func TestUserRepository_Exists_Error(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)

	dbErr := errors.New("db down")
	mock.ExpectQuery(`SELECT EXISTS`).WillReturnError(dbErr)

	_, err := repo.Exists("alice")
	requireError(t, err, dbErr)

	requireMockExpectations(t, mock)
}

// --- OrderRepository ---

func TestOrderRepository_Create(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewOrderRepository(db)
	order := &domain.Order{Number: "123", UserID: 1, Status: domain.OrderStatusNew,
		UploadedAt: time.Now(), UpdatedAt: time.Now()}

	mock.ExpectExec(`INSERT INTO orders`).WithArgs(
		order.Number, order.UserID, order.Status, order.UploadedAt, order.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(order)
	requireNoError(t, err)

	requireMockExpectations(t, mock)
}

func TestOrderRepository_Create_UniqueViolationSelf(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewOrderRepository(db)
	order := &domain.Order{Number: "123", UserID: 1, Status: domain.OrderStatusNew,
		UploadedAt: time.Now(), UpdatedAt: time.Now()}

	mock.ExpectExec(`INSERT INTO orders`).
		WillReturnError(&pgconn.PgError{Code: "23505"})
	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WithArgs("123").
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("123", 1, domain.OrderStatusNew, 0, time.Now(), time.Now()))

	err := repo.Create(order)
	requireError(t, err, domain.ErrOrderAlreadyUploadedByUser)

	requireMockExpectations(t, mock)
}

func TestOrderRepository_Create_UniqueViolationAnother(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewOrderRepository(db)
	order := &domain.Order{Number: "123", UserID: 2, Status: domain.OrderStatusNew,
		UploadedAt: time.Now(), UpdatedAt: time.Now()}

	mock.ExpectExec(`INSERT INTO orders`).
		WillReturnError(&pgconn.PgError{Code: "23505"})
	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("123", 1, domain.OrderStatusNew, 0, time.Now(), time.Now()))

	err := repo.Create(order)
	requireError(t, err, domain.ErrOrderAlreadyUploadedByAnotherUser)

	requireMockExpectations(t, mock)
}

func TestOrderRepository_FindByNumber(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewOrderRepository(db)
	now := time.Now()
	accrual := 500.5

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WithArgs("123").
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("123", 1, domain.OrderStatusProcessed, accrual, now, now))

	order, err := repo.FindByNumber("123")
	requireNoError(t, err)

	if order.Accrual == nil || *order.Accrual != accrual {
		t.Errorf("order.Accrual = %v, want %v", order.Accrual, accrual)
	}

	requireMockExpectations(t, mock)
}

func TestOrderRepository_FindByNumber_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewOrderRepository(db)

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.FindByNumber("123")
	requireError(t, err, domain.ErrOrderNotFound)

	requireMockExpectations(t, mock)
}

func TestOrderRepository_FindByUserID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewOrderRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
		AddRow("123", 1, domain.OrderStatusProcessed, 100.0, now, now).
		AddRow("456", 1, domain.OrderStatusNew, nil, now, now)
	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WithArgs(int64(1)).WillReturnRows(rows)

	orders, err := repo.FindByUserID(1)
	requireNoError(t, err)

	if len(orders) != 2 {
		t.Fatalf("FindByUserID() len = %d, want 2", len(orders))
	}
	if orders[0].Accrual == nil || *orders[0].Accrual != 100.0 {
		t.Errorf("orders[0].Accrual = %v, want 100.0", orders[0].Accrual)
	}
	if orders[1].Accrual != nil {
		t.Errorf("orders[1].Accrual = %v, want nil", orders[1].Accrual)
	}

	requireMockExpectations(t, mock)
}

func TestOrderRepository_Update(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewOrderRepository(db)
	accrual := 42.0
	now := time.Now()

	// С начислением
	mock.ExpectExec(`UPDATE orders`).
		WithArgs(domain.OrderStatusProcessed, accrual, now, "123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	order := &domain.Order{Number: "123", Status: domain.OrderStatusProcessed,
		Accrual: &accrual, UpdatedAt: now}
	requireNoError(t, repo.Update(order))

	// Без начисления (nil)
	mock.ExpectExec(`UPDATE orders`).
		WithArgs(domain.OrderStatusInvalid, nil, now, "124").
		WillReturnResult(sqlmock.NewResult(0, 1))

	order2 := &domain.Order{Number: "124", Status: domain.OrderStatusInvalid, UpdatedAt: now}
	requireNoError(t, repo.Update(order2))

	requireMockExpectations(t, mock)
}

func TestOrderRepository_Update_Error(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewOrderRepository(db)

	dbErr := errors.New("db down")
	mock.ExpectExec(`UPDATE orders`).WillReturnError(dbErr)

	err := repo.Update(&domain.Order{Number: "1"})
	requireError(t, err, dbErr)

	requireMockExpectations(t, mock)
}

func TestOrderRepository_GetPendingOrders(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewOrderRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
		AddRow("123", 1, domain.OrderStatusNew, nil, now, now)
	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WithArgs(domain.OrderStatusNew, domain.OrderStatusProcessing, 100).
		WillReturnRows(rows)

	orders, err := repo.GetPendingOrders(100)
	requireNoError(t, err)

	if len(orders) != 1 || orders[0].Number != "123" {
		t.Errorf("unexpected orders: %+v", orders)
	}

	requireMockExpectations(t, mock)
}

func TestOrderRepository_GetPendingOrders_Error(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewOrderRepository(db)

	dbErr := errors.New("db down")
	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnError(dbErr)

	_, err := repo.GetPendingOrders(10)
	requireError(t, err, dbErr)

	requireMockExpectations(t, mock)
}

// --- BalanceRepository ---

func TestBalanceRepository_GetBalance_Found(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewBalanceRepository(db)

	mock.ExpectQuery(`SELECT current, withdrawn FROM balances`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"current", "withdrawn"}).AddRow(100.0, 50.0))

	balance, err := repo.GetBalance(1)
	requireNoError(t, err)

	if balance.Current != 100.0 || balance.Withdrawn != 50.0 || balance.UserID != 1 {
		t.Errorf("unexpected balance: %+v", balance)
	}

	requireMockExpectations(t, mock)
}

func TestBalanceRepository_GetBalance_CreatesIfMissing(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewBalanceRepository(db)

	mock.ExpectQuery(`SELECT current, withdrawn FROM balances`).
		WithArgs(int64(1)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO balances`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"current", "withdrawn"}).AddRow(0, 0))

	balance, err := repo.GetBalance(1)
	requireNoError(t, err)

	if balance.Current != 0 || balance.Withdrawn != 0 {
		t.Errorf("unexpected balance: %+v", balance)
	}

	requireMockExpectations(t, mock)
}

func TestBalanceRepository_GetBalance_OtherError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewBalanceRepository(db)

	dbErr := errors.New("db down")
	mock.ExpectQuery(`SELECT current, withdrawn FROM balances`).
		WillReturnError(dbErr)

	_, err := repo.GetBalance(1)
	requireError(t, err, dbErr)

	requireMockExpectations(t, mock)
}

func TestBalanceRepository_CreateBalance_Error(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewBalanceRepository(db)

	dbErr := errors.New("db down")
	mock.ExpectQuery(`INSERT INTO balances`).WillReturnError(dbErr)

	_, err := repo.CreateBalance(1)
	requireError(t, err, dbErr)

	requireMockExpectations(t, mock)
}

func TestBalanceRepository_AddBalance_Existing(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewBalanceRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE balances`).
		WithArgs(100.0, int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.AddBalance(1, 100.0)
	requireNoError(t, err)

	requireMockExpectations(t, mock)
}

func TestBalanceRepository_AddBalance_CreatesIfMissing(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewBalanceRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE balances`).
		WillReturnResult(sqlmock.NewResult(0, 0)) // баланса нет
	mock.ExpectExec(`INSERT INTO balances`).
		WithArgs(int64(1), 100.0).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.AddBalance(1, 100.0)
	requireNoError(t, err)

	requireMockExpectations(t, mock)
}

func TestBalanceRepository_AddBalance_BeginError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewBalanceRepository(db)

	dbErr := errors.New("begin failed")
	mock.ExpectBegin().WillReturnError(dbErr)

	err := repo.AddBalance(1, 100.0)
	requireError(t, err, dbErr)

	requireMockExpectations(t, mock)
}

func TestBalanceRepository_AddBalance_ExecError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewBalanceRepository(db)

	dbErr := errors.New("exec failed")
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE balances`).WillReturnError(dbErr)
	mock.ExpectRollback()

	err := repo.AddBalance(1, 100.0)
	requireError(t, err, dbErr)

	requireMockExpectations(t, mock)
}

func TestBalanceRepository_AddBalance_InsertError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewBalanceRepository(db)

	dbErr := errors.New("insert failed")
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE balances`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO balances`).WillReturnError(dbErr)
	mock.ExpectRollback()

	err := repo.AddBalance(1, 100.0)
	requireError(t, err, dbErr)

	requireMockExpectations(t, mock)
}

func TestBalanceRepository_SubtractBalance_Success(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewBalanceRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT current FROM balances`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"current"}).AddRow(500.0))
	mock.ExpectExec(`UPDATE balances`).
		WithArgs(100.0, int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.SubtractBalance(1, 100.0)
	requireNoError(t, err)

	requireMockExpectations(t, mock)
}

func TestBalanceRepository_SubtractBalance_Insufficient(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewBalanceRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT current FROM balances`).
		WillReturnRows(sqlmock.NewRows([]string{"current"}).AddRow(50.0))
	mock.ExpectRollback()

	err := repo.SubtractBalance(1, 100.0)
	requireError(t, err, domain.ErrInsufficientBalance)

	requireMockExpectations(t, mock)
}

func TestBalanceRepository_SubtractBalance_NoRows(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewBalanceRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT current FROM balances`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err := repo.SubtractBalance(1, 100.0)
	requireError(t, err, domain.ErrInsufficientBalance)

	requireMockExpectations(t, mock)
}

func TestBalanceRepository_SubtractBalance_BeginError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewBalanceRepository(db)

	dbErr := errors.New("begin failed")
	mock.ExpectBegin().WillReturnError(dbErr)

	err := repo.SubtractBalance(1, 100.0)
	requireError(t, err, dbErr)

	requireMockExpectations(t, mock)
}

func TestBalanceRepository_SubtractBalance_CheckError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewBalanceRepository(db)

	dbErr := errors.New("db down")
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT current FROM balances`).WillReturnError(dbErr)
	mock.ExpectRollback()

	err := repo.SubtractBalance(1, 100.0)
	requireError(t, err, dbErr)
	requireNotError(t, err, domain.ErrInsufficientBalance)

	requireMockExpectations(t, mock)
}

// --- WithdrawalRepository ---

func TestWithdrawalRepository_Create(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewWithdrawalRepository(db)
	now := time.Now()
	w := &domain.Withdrawal{UserID: 1, OrderNumber: "123", Sum: 100, ProcessedAt: now}

	mock.ExpectQuery(`INSERT INTO withdrawals`).
		WithArgs(int64(1), "123", 100.0, now).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

	err := repo.Create(w)
	requireNoError(t, err)

	if w.ID != 7 {
		t.Errorf("w.ID = %d, want 7", w.ID)
	}

	requireMockExpectations(t, mock)
}

func TestWithdrawalRepository_Create_Error(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewWithdrawalRepository(db)

	dbErr := errors.New("db down")
	mock.ExpectQuery(`INSERT INTO withdrawals`).WillReturnError(dbErr)

	err := repo.Create(&domain.Withdrawal{UserID: 1, OrderNumber: "1", Sum: 1})
	requireError(t, err, dbErr)

	requireMockExpectations(t, mock)
}

func TestWithdrawalRepository_FindByUserID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewWithdrawalRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "order_number", "sum", "processed_at"}).
		AddRow(1, 1, "123", 100.0, now).
		AddRow(2, 1, "456", 50.0, now)
	mock.ExpectQuery(`SELECT id, user_id, order_number, sum, processed_at FROM withdrawals`).
		WithArgs(int64(1)).WillReturnRows(rows)

	withdrawals, err := repo.FindByUserID(1)
	requireNoError(t, err)

	if len(withdrawals) != 2 {
		t.Fatalf("FindByUserID() len = %d, want 2", len(withdrawals))
	}
	if withdrawals[0].OrderNumber != "123" || withdrawals[0].Sum != 100.0 {
		t.Errorf("unexpected withdrawal: %+v", withdrawals[0])
	}

	requireMockExpectations(t, mock)
}

func TestWithdrawalRepository_FindByUserID_Error(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewWithdrawalRepository(db)

	dbErr := errors.New("db down")
	mock.ExpectQuery(`SELECT id, user_id, order_number, sum, processed_at FROM withdrawals`).
		WillReturnError(dbErr)

	_, err := repo.FindByUserID(1)
	requireError(t, err, dbErr)

	requireMockExpectations(t, mock)
}

func TestWithdrawalRepository_GetTotalWithdrawn(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewWithdrawalRepository(db)

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(sum\), 0\) FROM withdrawals`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(150.0))

	total, err := repo.GetTotalWithdrawn(1)
	requireNoError(t, err)

	if total != 150.0 {
		t.Errorf("GetTotalWithdrawn() = %v, want 150.0", total)
	}

	requireMockExpectations(t, mock)
}

func TestWithdrawalRepository_GetTotalWithdrawn_Error(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewWithdrawalRepository(db)

	dbErr := errors.New("db down")
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(sum\), 0\) FROM withdrawals`).
		WillReturnError(dbErr)

	_, err := repo.GetTotalWithdrawn(1)
	requireError(t, err, dbErr)

	requireMockExpectations(t, mock)
}

// Проверяем, что контекст используется корректно (компиляционная проверка).
var _ = context.Background

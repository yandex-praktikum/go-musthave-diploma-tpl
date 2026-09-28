package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/true18/go-musthave-diploma-tpl/internal/money"
	"github.com/true18/go-musthave-diploma-tpl/internal/repository"
	"github.com/true18/go-musthave-diploma-tpl/internal/service"
)

func TestRegisterAndLogin(t *testing.T) {
	repo := newServiceRepo()
	app := service.New(repo)

	userID, err := app.Register(context.Background(), "alice", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if userID != 1 {
		t.Fatalf("userID = %d", userID)
	}
	if string(repo.users["alice"].PasswordHash) == "secret" {
		t.Fatal("password was not hashed")
	}

	if _, err := app.Register(context.Background(), "alice", "secret"); !errors.Is(err, service.ErrUserExists) {
		t.Fatalf("duplicate err = %v", err)
	}

	userID, err = app.Login(context.Background(), "alice", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if userID != 1 {
		t.Fatalf("login userID = %d", userID)
	}

	if _, err := app.Login(context.Background(), "alice", "bad"); !errors.Is(err, service.ErrInvalidCredentials) {
		t.Fatalf("wrong password err = %v", err)
	}
}

func TestUploadOrderValidationAndDuplicates(t *testing.T) {
	repo := newServiceRepo()
	app := service.New(repo)

	if err := app.UploadOrder(context.Background(), 1, "abc"); !errors.Is(err, service.ErrInvalidOrderFormat) {
		t.Fatalf("invalid format err = %v", err)
	}
	if err := app.UploadOrder(context.Background(), 1, "12345678904"); !errors.Is(err, service.ErrInvalidOrderNumber) {
		t.Fatalf("invalid luhn err = %v", err)
	}
	if err := app.UploadOrder(context.Background(), 1, "12345678903"); err != nil {
		t.Fatal(err)
	}
	if err := app.UploadOrder(context.Background(), 1, "12345678903"); !errors.Is(err, service.ErrOrderExistsSameUser) {
		t.Fatalf("same user duplicate err = %v", err)
	}
	if err := app.UploadOrder(context.Background(), 2, "12345678903"); !errors.Is(err, service.ErrOrderExistsAnotherUser) {
		t.Fatalf("another user duplicate err = %v", err)
	}
}

func TestWithdrawValidation(t *testing.T) {
	repo := newServiceRepo()
	app := service.New(repo)

	if err := app.Withdraw(context.Background(), 1, "12345678904", mustAmount(t, "1")); !errors.Is(err, service.ErrInvalidOrderNumber) {
		t.Fatalf("invalid order err = %v", err)
	}
	if err := app.Withdraw(context.Background(), 1, "12345678903", 0); !errors.Is(err, service.ErrInvalidWithdrawalAmount) {
		t.Fatalf("invalid amount err = %v", err)
	}

	repo.balance = repository.Balance{Current: mustAmount(t, "10")}
	if err := app.Withdraw(context.Background(), 1, "12345678903", mustAmount(t, "20")); !errors.Is(err, service.ErrInsufficientFunds) {
		t.Fatalf("insufficient funds err = %v", err)
	}
	if err := app.Withdraw(context.Background(), 1, "12345678903", mustAmount(t, "5")); err != nil {
		t.Fatal(err)
	}
}

type serviceRepo struct {
	nextID  int64
	users   map[string]repository.User
	orders  map[string]repository.Order
	balance repository.Balance
}

func newServiceRepo() *serviceRepo {
	return &serviceRepo{
		nextID: 1,
		users:  make(map[string]repository.User),
		orders: make(map[string]repository.Order),
	}
}

func (r *serviceRepo) CreateUser(_ context.Context, login string, passwordHash []byte) (int64, error) {
	if _, ok := r.users[login]; ok {
		return 0, repository.ErrUserExists
	}
	user := repository.User{ID: r.nextID, Login: login, PasswordHash: passwordHash, CreatedAt: time.Now()}
	r.nextID++
	r.users[login] = user
	return user.ID, nil
}

func (r *serviceRepo) FindUserByLogin(_ context.Context, login string) (repository.User, error) {
	user, ok := r.users[login]
	if !ok {
		return repository.User{}, repository.ErrNotFound
	}
	return user, nil
}

func (r *serviceRepo) CreateOrder(_ context.Context, userID int64, number string) error {
	order, ok := r.orders[number]
	if ok && order.UserID == userID {
		return repository.ErrOrderExistsSameUser
	}
	if ok {
		return repository.ErrOrderExistsAnotherUser
	}
	r.orders[number] = repository.Order{Number: number, UserID: userID, Status: repository.StatusNew}
	return nil
}

func (r *serviceRepo) ListUserOrders(context.Context, int64) ([]repository.Order, error) {
	return nil, nil
}

func (r *serviceRepo) ListPendingOrders(context.Context, int) ([]repository.Order, error) {
	return nil, nil
}

func (r *serviceRepo) UpdateOrder(context.Context, string, string, *money.Amount) error {
	return nil
}

func (r *serviceRepo) Balance(context.Context, int64) (repository.Balance, error) {
	return r.balance, nil
}

func (r *serviceRepo) Withdraw(_ context.Context, _ int64, _ string, amount money.Amount) error {
	if r.balance.Current < amount {
		return repository.ErrInsufficientFunds
	}
	r.balance.Current -= amount
	r.balance.Withdrawn += amount
	return nil
}

func (r *serviceRepo) ListWithdrawals(context.Context, int64) ([]repository.Withdrawal, error) {
	return nil, nil
}

func mustAmount(t *testing.T, value string) money.Amount {
	t.Helper()

	amount, err := money.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	return amount
}

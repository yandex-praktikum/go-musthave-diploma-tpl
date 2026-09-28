package service_test

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/true18/go-musthave-diploma-tpl/internal/accrual"
	"github.com/true18/go-musthave-diploma-tpl/internal/money"
	"github.com/true18/go-musthave-diploma-tpl/internal/repository"
	"github.com/true18/go-musthave-diploma-tpl/internal/service"
)

func TestWorkerMapsAccrualStatuses(t *testing.T) {
	tests := []struct {
		name        string
		external    string
		wantStatus  string
		wantAccrual bool
	}{
		{name: "registered", external: accrual.StatusRegistered, wantStatus: repository.StatusProcessing},
		{name: "processing", external: accrual.StatusProcessing, wantStatus: repository.StatusProcessing},
		{name: "invalid", external: accrual.StatusInvalid, wantStatus: repository.StatusInvalid},
		{name: "processed", external: accrual.StatusProcessed, wantStatus: repository.StatusProcessed, wantAccrual: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newWorkerRepo()
			repo.orders["12345678903"] = repository.Order{
				Number:     "12345678903",
				UserID:     1,
				Status:     repository.StatusNew,
				UploadedAt: time.Now(),
			}
			amount := mustAmount(t, "500")
			client := &workerClient{results: map[string]accrual.Result{
				"12345678903": {Order: "12345678903", Status: tt.external, Accrual: &amount},
			}}

			service.NewWorker(repo, client, zap.NewNop()).PollOnce(context.Background())

			order := repo.orders["12345678903"]
			if order.Status != tt.wantStatus {
				t.Fatalf("status = %q, want %q", order.Status, tt.wantStatus)
			}
			if (order.Accrual != nil) != tt.wantAccrual {
				t.Fatalf("accrual presence = %v, want %v", order.Accrual != nil, tt.wantAccrual)
			}
		})
	}
}

func TestWorkerNoContentLeavesOrderPending(t *testing.T) {
	repo := newWorkerRepo()
	repo.orders["12345678903"] = repository.Order{Number: "12345678903", UserID: 1, Status: repository.StatusNew}
	client := &workerClient{results: map[string]accrual.Result{
		"12345678903": {NoContent: true},
	}}

	service.NewWorker(repo, client, zap.NewNop()).PollOnce(context.Background())

	if repo.orders["12345678903"].Status != repository.StatusNew {
		t.Fatalf("status = %q", repo.orders["12345678903"].Status)
	}
}

func TestWorkerDoesNotApplyProcessedOrderTwice(t *testing.T) {
	repo := newWorkerRepo()
	repo.orders["12345678903"] = repository.Order{Number: "12345678903", UserID: 1, Status: repository.StatusNew}
	amount := mustAmount(t, "500")
	client := &workerClient{results: map[string]accrual.Result{
		"12345678903": {Order: "12345678903", Status: accrual.StatusProcessed, Accrual: &amount},
	}}
	worker := service.NewWorker(repo, client, zap.NewNop())

	worker.PollOnce(context.Background())
	worker.PollOnce(context.Background())

	if client.calls != 1 {
		t.Fatalf("client calls = %d, want 1", client.calls)
	}
	balance, err := repo.Balance(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if balance.Current != amount {
		t.Fatalf("balance = %s, want %s", balance.Current, amount)
	}
}

func TestWorkerHonorsRateLimit(t *testing.T) {
	repo := newWorkerRepo()
	repo.orders["12345678903"] = repository.Order{Number: "12345678903", UserID: 1, Status: repository.StatusNew}
	client := &workerClient{
		err:    accrual.ErrRateLimited,
		result: accrual.Result{RetryAfter: time.Minute},
	}
	worker := service.NewWorker(repo, client, zap.NewNop())

	worker.PollOnce(context.Background())
	worker.PollOnce(context.Background())

	if client.calls != 1 {
		t.Fatalf("client calls = %d, want 1 because Retry-After is active", client.calls)
	}
}

type workerClient struct {
	results map[string]accrual.Result
	result  accrual.Result
	err     error
	calls   int
}

func (c *workerClient) Fetch(_ context.Context, number string) (accrual.Result, error) {
	c.calls++
	if c.err != nil {
		return c.result, c.err
	}
	result, ok := c.results[number]
	if !ok {
		return accrual.Result{}, errors.New("unexpected order")
	}
	return result, nil
}

type workerRepo struct {
	orders      map[string]repository.Order
	withdrawals map[int64][]repository.Withdrawal
}

func newWorkerRepo() *workerRepo {
	return &workerRepo{
		orders:      make(map[string]repository.Order),
		withdrawals: make(map[int64][]repository.Withdrawal),
	}
}

func (r *workerRepo) CreateUser(context.Context, string, []byte) (int64, error) {
	return 0, nil
}

func (r *workerRepo) FindUserByLogin(context.Context, string) (repository.User, error) {
	return repository.User{}, nil
}

func (r *workerRepo) CreateOrder(_ context.Context, userID int64, number string) error {
	r.orders[number] = repository.Order{Number: number, UserID: userID, Status: repository.StatusNew}
	return nil
}

func (r *workerRepo) ListUserOrders(context.Context, int64) ([]repository.Order, error) {
	return nil, nil
}

func (r *workerRepo) ListPendingOrders(_ context.Context, limit int) ([]repository.Order, error) {
	var orders []repository.Order
	for _, order := range r.orders {
		if order.Status == repository.StatusNew || order.Status == repository.StatusProcessing {
			orders = append(orders, order)
		}
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].UploadedAt.Before(orders[j].UploadedAt)
	})
	if len(orders) > limit {
		orders = orders[:limit]
	}
	return orders, nil
}

func (r *workerRepo) UpdateOrder(_ context.Context, number string, status string, accrual *money.Amount) error {
	order := r.orders[number]
	if order.Status == repository.StatusProcessed || order.Status == repository.StatusInvalid {
		return nil
	}
	order.Status = status
	order.Accrual = accrual
	r.orders[number] = order
	return nil
}

func (r *workerRepo) Balance(_ context.Context, userID int64) (repository.Balance, error) {
	var current money.Amount
	var withdrawn money.Amount
	for _, order := range r.orders {
		if order.UserID == userID && order.Status == repository.StatusProcessed && order.Accrual != nil {
			current += *order.Accrual
		}
	}
	for _, withdrawal := range r.withdrawals[userID] {
		withdrawn += withdrawal.Sum
	}
	return repository.Balance{Current: current - withdrawn, Withdrawn: withdrawn}, nil
}

func (r *workerRepo) Withdraw(context.Context, int64, string, money.Amount) error {
	return nil
}

func (r *workerRepo) ListWithdrawals(context.Context, int64) ([]repository.Withdrawal, error) {
	return nil, nil
}

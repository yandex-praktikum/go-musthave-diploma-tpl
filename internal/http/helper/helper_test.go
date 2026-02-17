package helper

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/anon-d/gophermarket/internal/domain"
	"github.com/anon-d/gophermarket/internal/repository"
)

func TestToRepositoryUser(t *testing.T) {
	uid := uuid.New()
	domainUser := &domain.User{
		ID:       uid.String(),
		Login:    "testuser",
		PassHash: "hashedpassword",
	}

	repoUser := ToRepositoryUser(domainUser)

	if repoUser.ID != uid {
		t.Errorf("ToRepositoryUser() ID = %v, want %v", repoUser.ID, uid)
	}
	if repoUser.Login != domainUser.Login {
		t.Errorf("ToRepositoryUser() Login = %v, want %v", repoUser.Login, domainUser.Login)
	}
	if repoUser.PassHash != domainUser.PassHash {
		t.Errorf("ToRepositoryUser() PassHash = %v, want %v", repoUser.PassHash, domainUser.PassHash)
	}
}

func TestToDomainUser(t *testing.T) {
	uid := uuid.New()
	repoUser := &repository.User{
		ID:       uid,
		Login:    "testuser",
		PassHash: "hashedpassword",
	}

	domainUser := ToDomainUser(repoUser)

	if domainUser.ID != uid.String() {
		t.Errorf("ToDomainUser() ID = %v, want %v", domainUser.ID, uid.String())
	}
	if domainUser.Login != repoUser.Login {
		t.Errorf("ToDomainUser() Login = %v, want %v", domainUser.Login, repoUser.Login)
	}
	if domainUser.PassHash != repoUser.PassHash {
		t.Errorf("ToDomainUser() PassHash = %v, want %v", domainUser.PassHash, repoUser.PassHash)
	}
}

func TestToRepositoryOrder(t *testing.T) {
	uid := uuid.New()
	now := time.Now()
	domainOrder := &domain.Order{
		ID:         1,
		Number:     "123456789",
		UserID:     uid.String(),
		Status:     domain.OrderStatusNew,
		Accrual:    100.50,
		UploadedAt: now,
	}

	repoOrder := ToRepositoryOrder(domainOrder)

	if repoOrder.ID != domainOrder.ID {
		t.Errorf("ToRepositoryOrder() ID = %v, want %v", repoOrder.ID, domainOrder.ID)
	}
	if repoOrder.Number != domainOrder.Number {
		t.Errorf("ToRepositoryOrder() Number = %v, want %v", repoOrder.Number, domainOrder.Number)
	}
	if repoOrder.UserID != uid {
		t.Errorf("ToRepositoryOrder() UserID = %v, want %v", repoOrder.UserID, uid)
	}
	if repoOrder.Status != string(domainOrder.Status) {
		t.Errorf("ToRepositoryOrder() Status = %v, want %v", repoOrder.Status, string(domainOrder.Status))
	}
	if repoOrder.Accrual != domainOrder.Accrual {
		t.Errorf("ToRepositoryOrder() Accrual = %v, want %v", repoOrder.Accrual, domainOrder.Accrual)
	}
}

func TestToDomainOrder(t *testing.T) {
	uid := uuid.New()
	now := time.Now()
	repoOrder := &repository.Order{
		ID:         1,
		Number:     "123456789",
		UserID:     uid,
		Status:     "NEW",
		Accrual:    100.50,
		UploadedAt: now,
	}

	domainOrder := ToDomainOrder(repoOrder)

	if domainOrder.ID != repoOrder.ID {
		t.Errorf("ToDomainOrder() ID = %v, want %v", domainOrder.ID, repoOrder.ID)
	}
	if domainOrder.Number != repoOrder.Number {
		t.Errorf("ToDomainOrder() Number = %v, want %v", domainOrder.Number, repoOrder.Number)
	}
	if domainOrder.UserID != uid.String() {
		t.Errorf("ToDomainOrder() UserID = %v, want %v", domainOrder.UserID, uid.String())
	}
	if domainOrder.Status != domain.OrderStatus(repoOrder.Status) {
		t.Errorf("ToDomainOrder() Status = %v, want %v", domainOrder.Status, domain.OrderStatus(repoOrder.Status))
	}
}

func TestToDomainOrders(t *testing.T) {
	uid := uuid.New()
	now := time.Now()
	repoOrders := []repository.Order{
		{ID: 1, Number: "111", UserID: uid, Status: "NEW", Accrual: 10, UploadedAt: now},
		{ID: 2, Number: "222", UserID: uid, Status: "PROCESSED", Accrual: 20, UploadedAt: now},
		{ID: 3, Number: "333", UserID: uid, Status: "INVALID", Accrual: 0, UploadedAt: now},
	}

	domainOrders := ToDomainOrders(repoOrders)

	if len(domainOrders) != len(repoOrders) {
		t.Errorf("ToDomainOrders() len = %v, want %v", len(domainOrders), len(repoOrders))
	}

	for i, order := range domainOrders {
		if order.Number != repoOrders[i].Number {
			t.Errorf("ToDomainOrders()[%d].Number = %v, want %v", i, order.Number, repoOrders[i].Number)
		}
	}
}

func TestToDomainOrdersEmpty(t *testing.T) {
	repoOrders := []repository.Order{}
	domainOrders := ToDomainOrders(repoOrders)

	if len(domainOrders) != 0 {
		t.Errorf("ToDomainOrders() len = %v, want 0", len(domainOrders))
	}
}

func TestToRepositoryBalance(t *testing.T) {
	uid := uuid.New()
	domainBalance := &domain.Balance{
		UserID:    uid.String(),
		Current:   500.25,
		Withdrawn: 100.75,
	}

	repoBalance := ToRepositoryBalance(domainBalance)

	if repoBalance.UserID != uid {
		t.Errorf("ToRepositoryBalance() UserID = %v, want %v", repoBalance.UserID, uid)
	}
	if repoBalance.Current != domainBalance.Current {
		t.Errorf("ToRepositoryBalance() Current = %v, want %v", repoBalance.Current, domainBalance.Current)
	}
	if repoBalance.Withdrawn != domainBalance.Withdrawn {
		t.Errorf("ToRepositoryBalance() Withdrawn = %v, want %v", repoBalance.Withdrawn, domainBalance.Withdrawn)
	}
}

func TestToDomainBalance(t *testing.T) {
	uid := uuid.New()
	repoBalance := &repository.Balance{
		UserID:    uid,
		Current:   500.25,
		Withdrawn: 100.75,
	}

	domainBalance := ToDomainBalance(repoBalance)

	if domainBalance.UserID != uid.String() {
		t.Errorf("ToDomainBalance() UserID = %v, want %v", domainBalance.UserID, uid.String())
	}
	if domainBalance.Current != repoBalance.Current {
		t.Errorf("ToDomainBalance() Current = %v, want %v", domainBalance.Current, repoBalance.Current)
	}
	if domainBalance.Withdrawn != repoBalance.Withdrawn {
		t.Errorf("ToDomainBalance() Withdrawn = %v, want %v", domainBalance.Withdrawn, repoBalance.Withdrawn)
	}
}

func TestToRepositoryWithdrawal(t *testing.T) {
	uid := uuid.New()
	now := time.Now()
	domainWithdrawal := &domain.Withdrawal{
		ID:          1,
		UserID:      uid.String(),
		OrderNumber: "123456789",
		Sum:         50.00,
		ProcessedAt: now,
	}

	repoWithdrawal := ToRepositoryWithdrawal(domainWithdrawal)

	if repoWithdrawal.ID != domainWithdrawal.ID {
		t.Errorf("ToRepositoryWithdrawal() ID = %v, want %v", repoWithdrawal.ID, domainWithdrawal.ID)
	}
	if repoWithdrawal.UserID != uid {
		t.Errorf("ToRepositoryWithdrawal() UserID = %v, want %v", repoWithdrawal.UserID, uid)
	}
	if repoWithdrawal.OrderNumber != domainWithdrawal.OrderNumber {
		t.Errorf("ToRepositoryWithdrawal() OrderNumber = %v, want %v", repoWithdrawal.OrderNumber, domainWithdrawal.OrderNumber)
	}
	if repoWithdrawal.Sum != domainWithdrawal.Sum {
		t.Errorf("ToRepositoryWithdrawal() Sum = %v, want %v", repoWithdrawal.Sum, domainWithdrawal.Sum)
	}
}

func TestToDomainWithdrawal(t *testing.T) {
	uid := uuid.New()
	now := time.Now()
	repoWithdrawal := &repository.Withdrawal{
		ID:          1,
		UserID:      uid,
		OrderNumber: "123456789",
		Sum:         50.00,
		ProcessedAt: now,
	}

	domainWithdrawal := ToDomainWithdrawal(repoWithdrawal)

	if domainWithdrawal.ID != repoWithdrawal.ID {
		t.Errorf("ToDomainWithdrawal() ID = %v, want %v", domainWithdrawal.ID, repoWithdrawal.ID)
	}
	if domainWithdrawal.UserID != uid.String() {
		t.Errorf("ToDomainWithdrawal() UserID = %v, want %v", domainWithdrawal.UserID, uid.String())
	}
	if domainWithdrawal.OrderNumber != repoWithdrawal.OrderNumber {
		t.Errorf("ToDomainWithdrawal() OrderNumber = %v, want %v", domainWithdrawal.OrderNumber, repoWithdrawal.OrderNumber)
	}
	if domainWithdrawal.Sum != repoWithdrawal.Sum {
		t.Errorf("ToDomainWithdrawal() Sum = %v, want %v", domainWithdrawal.Sum, repoWithdrawal.Sum)
	}
}

func TestToDomainWithdrawals(t *testing.T) {
	uid := uuid.New()
	now := time.Now()
	repoWithdrawals := []repository.Withdrawal{
		{ID: 1, UserID: uid, OrderNumber: "111", Sum: 10, ProcessedAt: now},
		{ID: 2, UserID: uid, OrderNumber: "222", Sum: 20, ProcessedAt: now},
	}

	domainWithdrawals := ToDomainWithdrawals(repoWithdrawals)

	if len(domainWithdrawals) != len(repoWithdrawals) {
		t.Errorf("ToDomainWithdrawals() len = %v, want %v", len(domainWithdrawals), len(repoWithdrawals))
	}

	for i, w := range domainWithdrawals {
		if w.OrderNumber != repoWithdrawals[i].OrderNumber {
			t.Errorf("ToDomainWithdrawals()[%d].OrderNumber = %v, want %v", i, w.OrderNumber, repoWithdrawals[i].OrderNumber)
		}
	}
}

func TestToDomainWithdrawalsEmpty(t *testing.T) {
	repoWithdrawals := []repository.Withdrawal{}
	domainWithdrawals := ToDomainWithdrawals(repoWithdrawals)

	if len(domainWithdrawals) != 0 {
		t.Errorf("ToDomainWithdrawals() len = %v, want 0", len(domainWithdrawals))
	}
}

func TestToRepositoryUserInvalidUUID(t *testing.T) {
	domainUser := &domain.User{
		ID:       "invalid-uuid",
		Login:    "testuser",
		PassHash: "hashedpassword",
	}

	repoUser := ToRepositoryUser(domainUser)

	// При невалидном UUID должен вернуться нулевой UUID
	if repoUser.ID != uuid.Nil {
		t.Errorf("ToRepositoryUser() with invalid UUID should return uuid.Nil, got %v", repoUser.ID)
	}
}

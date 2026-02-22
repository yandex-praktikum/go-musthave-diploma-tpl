package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/anon-d/gophermarket/internal/repository"
	"github.com/anon-d/gophermarket/internal/repository/postgres"
	jwtpkg "github.com/anon-d/gophermarket/pkg/jwt"
)

const (
	testNamespace = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	testSecret    = "test-secret"
)

var testLogger = zap.NewNop()

func newTestService(repo Repository) *GopherService {
	return NewGopherService(testNamespace, repo, testLogger, testSecret)
}

// --- NewGopherService ---

func TestNewGopherService(t *testing.T) {
	svc := newTestService(nil)
	if svc == nil {
		t.Fatal("NewGopherService() returned nil")
	}
	if svc.tokenDuration != 24*time.Hour {
		t.Errorf("tokenDuration = %v, want 24h", svc.tokenDuration)
	}
}

// --- RegisterUser ---

func TestRegisterUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	mockRepo.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(nil)

	svc := newTestService(mockRepo)
	token, err := svc.RegisterUser(context.Background(), "newuser", "password123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}

	// Проверяем что токен валидный
	parsed, err := jwtpkg.GetToken(token, []byte(testSecret))
	if err != nil {
		t.Fatalf("token is not valid: %v", err)
	}
	claims, err := jwtpkg.GetClaimsVerified(parsed)
	if err != nil {
		t.Fatalf("claims verification failed: %v", err)
	}

	space := uuid.MustParse(testNamespace)
	expectedUID := uuid.NewSHA1(space, []byte("newuser")).String()
	if claims["sub"] != expectedUID {
		t.Errorf("sub = %v, want %v", claims["sub"], expectedUID)
	}
}

func TestRegisterUser_UserExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	mockRepo.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(postgres.ErrUserExists)

	svc := newTestService(mockRepo)
	_, err := svc.RegisterUser(context.Background(), "existing", "pass")

	if !errors.Is(err, ErrUserExists) {
		t.Errorf("err = %v, want %v", err, ErrUserExists)
	}
}

func TestRegisterUser_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	dbErr := errors.New("connection refused")
	mockRepo.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(dbErr)

	svc := newTestService(mockRepo)
	_, err := svc.RegisterUser(context.Background(), "user", "pass")

	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, ErrUserExists) {
		t.Error("should not be ErrUserExists")
	}
}

// --- LoginUser ---

func TestLoginUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	passHash, _ := bcrypt.GenerateFromPassword([]byte("correct-pass"), bcrypt.MinCost)
	uid := uuid.New()

	mockRepo.EXPECT().
		GetUserByLogin(gomock.Any(), "user1").
		Return(&repository.User{
			ID:       uid,
			Login:    "user1",
			PassHash: string(passHash),
		}, nil)

	svc := newTestService(mockRepo)
	token, err := svc.LoginUser(context.Background(), "user1", "correct-pass")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestLoginUser_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	mockRepo.EXPECT().
		GetUserByLogin(gomock.Any(), "unknown").
		Return(nil, postgres.ErrUserNotFound)

	svc := newTestService(mockRepo)
	_, err := svc.LoginUser(context.Background(), "unknown", "pass")

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("err = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestLoginUser_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	passHash, _ := bcrypt.GenerateFromPassword([]byte("correct-pass"), bcrypt.MinCost)

	mockRepo.EXPECT().
		GetUserByLogin(gomock.Any(), "user1").
		Return(&repository.User{
			ID:       uuid.New(),
			Login:    "user1",
			PassHash: string(passHash),
		}, nil)

	svc := newTestService(mockRepo)
	_, err := svc.LoginUser(context.Background(), "user1", "wrong-pass")

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("err = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestLoginUser_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	mockRepo.EXPECT().
		GetUserByLogin(gomock.Any(), "user1").
		Return(nil, errors.New("db error"))

	svc := newTestService(mockRepo)
	_, err := svc.LoginUser(context.Background(), "user1", "pass")

	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, ErrInvalidCredentials) {
		t.Error("should not be ErrInvalidCredentials")
	}
}

// --- CreateOrder ---

func TestCreateOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	mockRepo.EXPECT().
		CreateOrder(gomock.Any(), gomock.Any()).
		Return(nil)

	svc := newTestService(mockRepo)
	// 12345678903 — валидный номер по Луну
	err := svc.CreateOrder(context.Background(), "user-1", "12345678903")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateOrder_InvalidLuhn(t *testing.T) {
	svc := newTestService(nil) // mock не нужен — до репозитория не дойдёт
	err := svc.CreateOrder(context.Background(), "user-1", "123")

	if !errors.Is(err, ErrInvalidOrderNumber) {
		t.Errorf("err = %v, want %v", err, ErrInvalidOrderNumber)
	}
}

func TestCreateOrder_OrderExistsSameUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	mockRepo.EXPECT().
		CreateOrder(gomock.Any(), gomock.Any()).
		Return(postgres.ErrOrderExists)

	svc := newTestService(mockRepo)
	err := svc.CreateOrder(context.Background(), "user-1", "12345678903")

	if !errors.Is(err, ErrOrderExists) {
		t.Errorf("err = %v, want %v", err, ErrOrderExists)
	}
}

func TestCreateOrder_OrderExistsAnotherUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	mockRepo.EXPECT().
		CreateOrder(gomock.Any(), gomock.Any()).
		Return(postgres.ErrOrderExistsByAnotherUser)

	svc := newTestService(mockRepo)
	err := svc.CreateOrder(context.Background(), "user-1", "12345678903")

	if !errors.Is(err, ErrOrderExistsByAnotherUser) {
		t.Errorf("err = %v, want %v", err, ErrOrderExistsByAnotherUser)
	}
}

func TestCreateOrder_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	mockRepo.EXPECT().
		CreateOrder(gomock.Any(), gomock.Any()).
		Return(errors.New("db error"))

	svc := newTestService(mockRepo)
	err := svc.CreateOrder(context.Background(), "user-1", "12345678903")

	if err == nil {
		t.Fatal("expected error")
	}
}

// --- GetOrders ---

func TestGetOrders_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	uid := uuid.New()
	now := time.Now()
	repoOrders := []repository.Order{
		{ID: 1, Number: "111", UserID: uid, Status: "NEW", Accrual: 0, UploadedAt: now},
		{ID: 2, Number: "222", UserID: uid, Status: "PROCESSED", Accrual: 500, UploadedAt: now},
	}

	mockRepo.EXPECT().
		GetOrdersByUserID(gomock.Any(), "user-1").
		Return(repoOrders, nil)

	svc := newTestService(mockRepo)
	orders, err := svc.GetOrders(context.Background(), "user-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 2 {
		t.Fatalf("len(orders) = %d, want 2", len(orders))
	}
	if orders[0].Number != "111" {
		t.Errorf("orders[0].Number = %q, want %q", orders[0].Number, "111")
	}
}

func TestGetOrders_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	mockRepo.EXPECT().
		GetOrdersByUserID(gomock.Any(), "user-1").
		Return(nil, errors.New("db error"))

	svc := newTestService(mockRepo)
	_, err := svc.GetOrders(context.Background(), "user-1")

	if err == nil {
		t.Fatal("expected error")
	}
}

// --- GetBalance ---

func TestGetBalance_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	uid := uuid.New()
	mockRepo.EXPECT().
		GetBalance(gomock.Any(), "user-1").
		Return(&repository.Balance{UserID: uid, Current: 100.5, Withdrawn: 50.0}, nil)

	svc := newTestService(mockRepo)
	balance, err := svc.GetBalance(context.Background(), "user-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance.Current != 100.5 {
		t.Errorf("Current = %v, want 100.5", balance.Current)
	}
	if balance.Withdrawn != 50.0 {
		t.Errorf("Withdrawn = %v, want 50.0", balance.Withdrawn)
	}
}

func TestGetBalance_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	mockRepo.EXPECT().
		GetBalance(gomock.Any(), "user-1").
		Return(nil, errors.New("db error"))

	svc := newTestService(mockRepo)
	_, err := svc.GetBalance(context.Background(), "user-1")

	if err == nil {
		t.Fatal("expected error")
	}
}

// --- Withdraw ---

func TestWithdraw_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	mockRepo.EXPECT().
		Withdraw(gomock.Any(), gomock.Any()).
		Return(nil)

	svc := newTestService(mockRepo)
	err := svc.Withdraw(context.Background(), "user-1", "12345678903", 100.0)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWithdraw_InvalidLuhn(t *testing.T) {
	svc := newTestService(nil)
	err := svc.Withdraw(context.Background(), "user-1", "123", 10.0)

	if !errors.Is(err, ErrInvalidOrderNumber) {
		t.Errorf("err = %v, want %v", err, ErrInvalidOrderNumber)
	}
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	mockRepo.EXPECT().
		Withdraw(gomock.Any(), gomock.Any()).
		Return(postgres.ErrInsufficientFunds)

	svc := newTestService(mockRepo)
	err := svc.Withdraw(context.Background(), "user-1", "12345678903", 9999.0)

	if !errors.Is(err, ErrInsufficientFunds) {
		t.Errorf("err = %v, want %v", err, ErrInsufficientFunds)
	}
}

func TestWithdraw_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	mockRepo.EXPECT().
		Withdraw(gomock.Any(), gomock.Any()).
		Return(errors.New("db error"))

	svc := newTestService(mockRepo)
	err := svc.Withdraw(context.Background(), "user-1", "12345678903", 10.0)

	if err == nil {
		t.Fatal("expected error")
	}
}

// --- GetWithdrawals ---

func TestGetWithdrawals_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	uid := uuid.New()
	now := time.Now()
	repoWithdrawals := []repository.Withdrawal{
		{ID: 1, UserID: uid, OrderNumber: "111", Sum: 50.0, ProcessedAt: now},
		{ID: 2, UserID: uid, OrderNumber: "222", Sum: 25.0, ProcessedAt: now},
	}

	mockRepo.EXPECT().
		GetWithdrawals(gomock.Any(), "user-1").
		Return(repoWithdrawals, nil)

	svc := newTestService(mockRepo)
	withdrawals, err := svc.GetWithdrawals(context.Background(), "user-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(withdrawals) != 2 {
		t.Fatalf("len(withdrawals) = %d, want 2", len(withdrawals))
	}
	if withdrawals[0].OrderNumber != "111" {
		t.Errorf("withdrawals[0].OrderNumber = %q, want %q", withdrawals[0].OrderNumber, "111")
	}
}

func TestGetWithdrawals_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)

	mockRepo.EXPECT().
		GetWithdrawals(gomock.Any(), "user-1").
		Return(nil, errors.New("db error"))

	svc := newTestService(mockRepo)
	_, err := svc.GetWithdrawals(context.Background(), "user-1")

	if err == nil {
		t.Fatal("expected error")
	}
}

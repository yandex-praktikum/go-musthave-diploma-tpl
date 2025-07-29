package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vglushak/go-musthave-diploma-tpl/internal/models"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/services"
)

// MockStorage для бенчмарков обработчиков
type MockHandlerStorage struct{}

func (m *MockHandlerStorage) Close() error {
	return nil
}

func (m *MockHandlerStorage) Ping(ctx context.Context) error {
	return nil
}

func (m *MockHandlerStorage) CreateUser(ctx context.Context, login, passwordHash string) (*models.User, error) {
	return &models.User{
		ID:       1,
		Login:    login,
		Password: passwordHash,
	}, nil
}

func (m *MockHandlerStorage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	if login == "existinguser" {
		return &models.User{
			ID:       1,
			Login:    login,
			Password: "$2a$10$hashedpassword",
		}, nil
	}
	return nil, nil
}

func (m *MockHandlerStorage) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return &models.User{
		ID:       id,
		Login:    "testuser",
		Password: "hashed_password",
	}, nil
}

func (m *MockHandlerStorage) CreateOrder(ctx context.Context, userID int64, number string) (*models.Order, error) {
	accrual := 0.0
	return &models.Order{
		ID:         1,
		UserID:     userID,
		Number:     number,
		Status:     "NEW",
		Accrual:    &accrual,
		UploadedAt: time.Now(),
	}, nil
}

func (m *MockHandlerStorage) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	return nil, nil
}

func (m *MockHandlerStorage) GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error) {
	accrual1 := 100.5
	accrual2 := 0.0
	now := time.Now()
	return []models.Order{
		{
			ID:         1,
			UserID:     userID,
			Number:     "1234567890",
			Status:     "PROCESSED",
			Accrual:    &accrual1,
			UploadedAt: now,
		},
		{
			ID:         2,
			UserID:     userID,
			Number:     "0987654321",
			Status:     "NEW",
			Accrual:    &accrual2,
			UploadedAt: now,
		},
	}, nil
}

func (m *MockHandlerStorage) UpdateOrderStatus(ctx context.Context, number string, status string, accrual *float64) error {
	return nil
}

func (m *MockHandlerStorage) GetBalance(ctx context.Context, userID int64) (*models.Balance, error) {
	return &models.Balance{
		UserID:    userID,
		Current:   1000.0,
		Withdrawn: 500.0,
	}, nil
}

func (m *MockHandlerStorage) UpdateBalance(ctx context.Context, userID int64, current, withdrawn float64) error {
	return nil
}

func (m *MockHandlerStorage) CreateWithdrawal(ctx context.Context, userID int64, order string, sum float64) (*models.Withdrawal, error) {
	return &models.Withdrawal{
		ID:          1,
		UserID:      userID,
		Order:       order,
		Sum:         sum,
		ProcessedAt: time.Now(),
	}, nil
}

func (m *MockHandlerStorage) GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	return []models.Withdrawal{
		{
			ID:          1,
			UserID:      userID,
			Order:       "1234567890",
			Sum:         100.0,
			ProcessedAt: time.Now(),
		},
	}, nil
}

func (m *MockHandlerStorage) GetOrdersByStatus(ctx context.Context, statuses []string) ([]models.Order, error) {
	accrual := 0.0
	now := time.Now()
	return []models.Order{
		{
			ID:         1,
			UserID:     1,
			Number:     "1234567890",
			Status:     "NEW",
			Accrual:    &accrual,
			UploadedAt: now,
		},
	}, nil
}

func (m *MockHandlerStorage) GetOrdersByStatusPaginated(ctx context.Context, statuses []string, limit, offset int) ([]models.Order, error) {
	accrual := 0.0
	now := time.Now()
	return []models.Order{
		{
			ID:         1,
			UserID:     1,
			Number:     "1234567890",
			Status:     "NEW",
			Accrual:    &accrual,
			UploadedAt: now,
		},
	}, nil
}

func (m *MockHandlerStorage) ProcessWithdrawal(ctx context.Context, userID int64, order string, sum float64) (*models.Withdrawal, error) {
	return &models.Withdrawal{
		ID:          1,
		UserID:      userID,
		Order:       order,
		Sum:         sum,
		ProcessedAt: time.Now(),
	}, nil
}

func (m *MockHandlerStorage) UpdateOrderStatusAndBalance(ctx context.Context, orderNumber string, status string, accrual *float64, userID int64, newCurrent, withdrawn float64) error {
	return nil
}

// Бенчмарки
func BenchmarkValidateOrderNumber(b *testing.B) {
	testNumbers := []string{
		"1234567890",
		"0987654321",
		"1234567890123456",
		"0000000000",
		"1111111111",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		number := testNumbers[i%len(testNumbers)]
		_ = validateOrderNumber(number)
	}
}

func BenchmarkRegisterHandler(b *testing.B) {
	storage := &MockHandlerStorage{}
	authService := services.NewAuthService("test-secret")
	accrualService := services.NewAccrualService("http://localhost:8080")

	handlers := NewHandlers(storage, authService, accrualService)

	reqBody := models.UserRegisterRequest{
		Login:    "newuser",
		Password: "password123",
	}

	jsonBody, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.RegisterHandler(w, req)
	}
}

func BenchmarkLoginHandler(b *testing.B) {
	storage := &MockHandlerStorage{}
	authService := services.NewAuthService("test-secret")
	accrualService := services.NewAccrualService("http://localhost:8080")

	handlers := NewHandlers(storage, authService, accrualService)

	reqBody := models.UserLoginRequest{
		Login:    "existinguser",
		Password: "password123",
	}

	jsonBody, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/user/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.LoginHandler(w, req)
	}
}

func BenchmarkUploadOrderHandler(b *testing.B) {
	storage := &MockHandlerStorage{}
	authService := services.NewAuthService("test-secret")
	accrualService := services.NewAccrualService("http://localhost:8080")

	handlers := NewHandlers(storage, authService, accrualService)

	orderNumber := "1234567890"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/user/orders", bytes.NewBufferString(orderNumber))
		req.Header.Set("Content-Type", "text/plain")
		ctx := context.WithValue(req.Context(), "userID", int64(1))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handlers.UploadOrderHandler(w, req)
	}
}

func BenchmarkGetOrdersHandler(b *testing.B) {
	storage := &MockHandlerStorage{}
	authService := services.NewAuthService("test-secret")
	accrualService := services.NewAccrualService("http://localhost:8080")

	handlers := NewHandlers(storage, authService, accrualService)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/user/orders", nil)
		ctx := context.WithValue(req.Context(), "userID", int64(1))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handlers.GetOrdersHandler(w, req)
	}
}

func BenchmarkGetBalanceHandler(b *testing.B) {
	storage := &MockHandlerStorage{}
	authService := services.NewAuthService("test-secret")
	accrualService := services.NewAccrualService("http://localhost:8080")

	handlers := NewHandlers(storage, authService, accrualService)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/user/balance", nil)
		ctx := context.WithValue(req.Context(), "userID", int64(1))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handlers.GetBalanceHandler(w, req)
	}
}

func BenchmarkWithdrawHandler(b *testing.B) {
	storage := &MockHandlerStorage{}
	authService := services.NewAuthService("test-secret")
	accrualService := services.NewAccrualService("http://localhost:8080")

	handlers := NewHandlers(storage, authService, accrualService)

	reqBody := models.WithdrawRequest{
		Order: "1234567890",
		Sum:   100.0,
	}

	jsonBody, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/user/balance/withdraw", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), "userID", int64(1))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handlers.WithdrawHandler(w, req)
	}
}

func BenchmarkGetWithdrawalsHandler(b *testing.B) {
	storage := &MockHandlerStorage{}
	authService := services.NewAuthService("test-secret")
	accrualService := services.NewAccrualService("http://localhost:8080")

	handlers := NewHandlers(storage, authService, accrualService)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/user/withdrawals", nil)
		ctx := context.WithValue(req.Context(), "userID", int64(1))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handlers.GetWithdrawalsHandler(w, req)
	}
}

// Бенчмарк для JSON маршалинга/анмаршалинга
func BenchmarkJSONMarshal(b *testing.B) {
	user := models.User{
		ID:       1,
		Login:    "testuser",
		Password: "hashed_password",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(user)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONUnmarshal(b *testing.B) {
	jsonData := []byte(`{"login":"testuser","password":"password123"}`)
	var req models.UserRegisterRequest

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := json.Unmarshal(jsonData, &req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

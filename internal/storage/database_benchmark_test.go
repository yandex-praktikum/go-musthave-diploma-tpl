package storage

import (
	"context"
	"testing"
	"time"

	"github.com/vglushak/go-musthave-diploma-tpl/internal/models"
)

// MockStorage для бенчмарков
type MockStorage struct{}

func (m *MockStorage) Close() error {
	return nil
}

func (m *MockStorage) Ping(ctx context.Context) error {
	return nil
}

func (m *MockStorage) CreateUser(ctx context.Context, login, passwordHash string) (*models.User, error) {
	return &models.User{
		ID:       1,
		Login:    login,
		Password: passwordHash,
	}, nil
}

func (m *MockStorage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	return &models.User{
		ID:       1,
		Login:    login,
		Password: "hashed_password",
	}, nil
}

func (m *MockStorage) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return &models.User{
		ID:       id,
		Login:    "testuser",
		Password: "hashed_password",
	}, nil
}

func (m *MockStorage) CreateOrder(ctx context.Context, userID int64, number string) (*models.Order, error) {
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

func (m *MockStorage) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	accrual := 100.5
	return &models.Order{
		ID:         1,
		UserID:     1,
		Number:     number,
		Status:     "PROCESSED",
		Accrual:    &accrual,
		UploadedAt: time.Now(),
	}, nil
}

func (m *MockStorage) GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error) {
	accrual1 := 100.5
	accrual2 := 0.0
	return []models.Order{
		{
			ID:         1,
			UserID:     userID,
			Number:     "1234567890",
			Status:     "PROCESSED",
			Accrual:    &accrual1,
			UploadedAt: time.Now(),
		},
		{
			ID:         2,
			UserID:     userID,
			Number:     "0987654321",
			Status:     "NEW",
			Accrual:    &accrual2,
			UploadedAt: time.Now(),
		},
	}, nil
}

func (m *MockStorage) UpdateOrderStatus(ctx context.Context, number string, status string, accrual *float64) error {
	return nil
}

func (m *MockStorage) GetBalance(ctx context.Context, userID int64) (*models.Balance, error) {
	return &models.Balance{
		UserID:    userID,
		Current:   1000.0,
		Withdrawn: 500.0,
	}, nil
}

func (m *MockStorage) UpdateBalance(ctx context.Context, userID int64, current, withdrawn float64) error {
	return nil
}

func (m *MockStorage) CreateWithdrawal(ctx context.Context, userID int64, order string, sum float64) (*models.Withdrawal, error) {
	return &models.Withdrawal{
		ID:          1,
		UserID:      userID,
		Order:       order,
		Sum:         sum,
		ProcessedAt: time.Now(),
	}, nil
}

func (m *MockStorage) GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
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

func (m *MockStorage) GetOrdersByStatus(ctx context.Context, statuses []string) ([]models.Order, error) {
	accrual := 0.0
	return []models.Order{
		{
			ID:         1,
			UserID:     1,
			Number:     "1234567890",
			Status:     "NEW",
			Accrual:    &accrual,
			UploadedAt: time.Now(),
		},
	}, nil
}

func (m *MockStorage) GetOrdersByStatusPaginated(ctx context.Context, statuses []string, limit, offset int) ([]models.Order, error) {
	accrual := 0.0
	return []models.Order{
		{
			ID:         1,
			UserID:     1,
			Number:     "1234567890",
			Status:     "NEW",
			Accrual:    &accrual,
			UploadedAt: time.Now(),
		},
	}, nil
}

func (m *MockStorage) ProcessWithdrawal(ctx context.Context, userID int64, order string, sum float64) (*models.Withdrawal, error) {
	return &models.Withdrawal{
		ID:          1,
		UserID:      userID,
		Order:       order,
		Sum:         sum,
		ProcessedAt: time.Now(),
	}, nil
}

func (m *MockStorage) UpdateOrderStatusAndBalance(ctx context.Context, orderNumber string, status string, accrual *float64, userID int64, newCurrent, withdrawn float64) error {
	return nil
}

// Бенчмарки
func BenchmarkCreateUser(b *testing.B) {
	storage := &MockStorage{}
	ctx := context.Background()
	login := "testuser"
	passwordHash := "hashed_password"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.CreateUser(ctx, login, passwordHash)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetUserByLogin(b *testing.B) {
	storage := &MockStorage{}
	ctx := context.Background()
	login := "testuser"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetUserByLogin(ctx, login)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetUserByID(b *testing.B) {
	storage := &MockStorage{}
	ctx := context.Background()
	userID := int64(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetUserByID(ctx, userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateOrder(b *testing.B) {
	storage := &MockStorage{}
	ctx := context.Background()
	userID := int64(1)
	number := "1234567890"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.CreateOrder(ctx, userID, number)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetOrderByNumber(b *testing.B) {
	storage := &MockStorage{}
	ctx := context.Background()
	number := "1234567890"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetOrderByNumber(ctx, number)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetOrdersByUserID(b *testing.B) {
	storage := &MockStorage{}
	ctx := context.Background()
	userID := int64(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetOrdersByUserID(ctx, userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateOrderStatus(b *testing.B) {
	storage := &MockStorage{}
	ctx := context.Background()
	number := "1234567890"
	status := "PROCESSED"
	accrual := 100.5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := storage.UpdateOrderStatus(ctx, number, status, &accrual)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetBalance(b *testing.B) {
	storage := &MockStorage{}
	ctx := context.Background()
	userID := int64(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetBalance(ctx, userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateBalance(b *testing.B) {
	storage := &MockStorage{}
	ctx := context.Background()
	userID := int64(1)
	current := 1000.0
	withdrawn := 500.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := storage.UpdateBalance(ctx, userID, current, withdrawn)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateWithdrawal(b *testing.B) {
	storage := &MockStorage{}
	ctx := context.Background()
	userID := int64(1)
	order := "1234567890"
	sum := 100.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.CreateWithdrawal(ctx, userID, order, sum)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetWithdrawalsByUserID(b *testing.B) {
	storage := &MockStorage{}
	ctx := context.Background()
	userID := int64(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetWithdrawalsByUserID(ctx, userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetOrdersByStatus(b *testing.B) {
	storage := &MockStorage{}
	ctx := context.Background()
	statuses := []string{"NEW", "PROCESSING"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetOrdersByStatus(ctx, statuses)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetOrdersByStatusPaginated(b *testing.B) {
	storage := &MockStorage{}
	ctx := context.Background()
	statuses := []string{"NEW", "PROCESSING"}
	limit := 10
	offset := 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetOrdersByStatusPaginated(ctx, statuses, limit, offset)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProcessWithdrawal(b *testing.B) {
	storage := &MockStorage{}
	ctx := context.Background()
	userID := int64(1)
	order := "1234567890"
	sum := 100.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.ProcessWithdrawal(ctx, userID, order, sum)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateOrderStatusAndBalance(b *testing.B) {
	storage := &MockStorage{}
	ctx := context.Background()
	orderNumber := "1234567890"
	status := "PROCESSED"
	accrual := 100.5
	userID := int64(1)
	newCurrent := 1100.5
	withdrawn := 500.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := storage.UpdateOrderStatusAndBalance(ctx, orderNumber, status, &accrual, userID, newCurrent, withdrawn)
		if err != nil {
			b.Fatal(err)
		}
	}
}

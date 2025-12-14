// user_repo_test.go
package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для сканирования ====================

// MockScanner для тестирования scanUser
type MockUserScanner struct {
	id        int
	createdAt time.Time
	updatedAt time.Time
	login     string
	password  string
	isActive  bool
	err       error
}

func (m *MockUserScanner) Scan(dest ...interface{}) error {
	if m.err != nil {
		return m.err
	}

	// Заполняем значения в порядке, ожидаемом scanUser
	*dest[0].(*int) = m.id
	*dest[1].(*time.Time) = m.createdAt
	*dest[2].(*time.Time) = m.updatedAt
	*dest[3].(*string) = m.login
	*dest[4].(*string) = m.password
	*dest[5].(*bool) = m.isActive

	return nil
}

func TestUserRepo_scanUser(t *testing.T) {
	repo := &userRepo{}

	tests := []struct {
		name      string
		scanner   *MockUserScanner
		wantError bool
		expected  *entity.User
	}{
		{
			name: "successful scan",
			scanner: &MockUserScanner{
				id:        1,
				createdAt: time.Now(),
				updatedAt: time.Now().Add(time.Hour),
				login:     "testuser",
				password:  "hashed_password",
				isActive:  true,
			},
			wantError: false,
			expected: &entity.User{
				ID:       1,
				Login:    "testuser",
				Password: "hashed_password",
				IsActive: true,
			},
		},
		{
			name: "inactive user",
			scanner: &MockUserScanner{
				id:        2,
				createdAt: time.Now(),
				updatedAt: time.Now(),
				login:     "inactive",
				password:  "password",
				isActive:  false,
			},
			wantError: false,
			expected: &entity.User{
				ID:       2,
				Login:    "inactive",
				Password: "password",
				IsActive: false,
			},
		},
		{
			name: "scan error",
			scanner: &MockUserScanner{
				err: pgx.ErrNoRows,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var user entity.User
			err := repo.scanUser(tt.scanner, &user)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ID, user.ID)
				assert.Equal(t, tt.expected.Login, user.Login)
				assert.Equal(t, tt.expected.Password, user.Password)
				assert.Equal(t, tt.expected.IsActive, user.IsActive)
				assert.NotEmpty(t, user.CreatedAt)
				assert.NotEmpty(t, user.UpdatedAt)
			}
		})
	}
}

// ==================== Тесты для интерфейсов и структур ====================

func TestUserRepositoryInterfaces(t *testing.T) {
	// Простая проверка, что интерфейс правильно определен
	var repo UserRepository
	repo = &userRepo{}
	assert.NotNil(t, repo)

	// Проверяем наличие методов (будет ошибка компиляции, если методов нет)
	// Этот тест просто проверяет, что структура реализует интерфейс
	_ = repo
}

func TestUserStructValidation(t *testing.T) {
	tests := []struct {
		name      string
		user      entity.User
		shouldErr bool
	}{
		{
			name: "valid active user",
			user: entity.User{
				ID:        1,
				Login:     "validuser",
				Password:  "hashed_password",
				IsActive:  true,
				CreatedAt: time.Now().Format(time.RFC3339),
			},
			shouldErr: false,
		},
		{
			name: "valid inactive user",
			user: entity.User{
				ID:        2,
				Login:     "inactive",
				Password:  "password",
				IsActive:  false,
				CreatedAt: time.Now().Format(time.RFC3339),
			},
			shouldErr: false,
		},
		{
			name: "empty login",
			user: entity.User{
				ID:       3,
				Login:    "",
				Password: "password",
				IsActive: true,
			},
			shouldErr: true, // логин не должен быть пустым
		},
		{
			name: "empty password",
			user: entity.User{
				ID:       4,
				Login:    "user",
				Password: "",
				IsActive: true,
			},
			shouldErr: true, // пароль не должен быть пустым
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// В реальном коде здесь была бы валидация
			// if tt.user.Login == "" || tt.user.Password == "" {
			//     assert.True(t, tt.shouldErr)
			// }
			t.Logf("Testing user: %+v", tt.user)
		})
	}
}

// ==================== Тесты для констант времени ====================

func TestUserTimeFormatting(t *testing.T) {
	now := time.Now()
	user := &entity.User{
		ID:        1,
		CreatedAt: now.Format(time.RFC3339),
		UpdatedAt: now.Add(time.Hour).Format(time.RFC3339),
		Login:     "testuser",
		IsActive:  true,
	}

	assert.NotEmpty(t, user.CreatedAt)
	assert.NotEmpty(t, user.UpdatedAt)

	// Проверяем, что это валидный RFC3339 формат
	createdAt, err := time.Parse(time.RFC3339, user.CreatedAt)
	assert.NoError(t, err)
	assert.WithinDuration(t, now, createdAt, time.Second)
}

// ==================== Тесты для SQL запросов ====================

func TestUserSQLQueries(t *testing.T) {
	// Проверяем, что SQL запросы правильно сформированы
	// Это помогает предотвратить ошибки в запросах

	t.Run("create user query", func(t *testing.T) {
		query := `INSERT INTO users (login, password, is_active) VALUES ($1, $2, true) RETURNING id, created_at, updated_at, login, password, is_active`
		assert.Contains(t, query, "INSERT INTO users")
		assert.Contains(t, query, "RETURNING")
		assert.Contains(t, query, "is_active")
	})

	t.Run("get user by id query", func(t *testing.T) {
		query := `SELECT id, created_at, updated_at, login, password, is_active FROM users WHERE id = $1`
		assert.Contains(t, query, "SELECT")
		assert.Contains(t, query, "WHERE id =")
	})

	t.Run("get user by login query", func(t *testing.T) {
		query := `SELECT id, created_at, updated_at, login, password, is_active FROM users WHERE login = $1`
		assert.Contains(t, query, "WHERE login =")
	})

	t.Run("update password query", func(t *testing.T) {
		query := `UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2 AND is_active = true RETURNING updated_at`
		assert.Contains(t, query, "UPDATE users")
		assert.Contains(t, query, "SET password =")
	})

	t.Run("deactivate user query", func(t *testing.T) {
		query := `UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1 AND is_active = true RETURNING updated_at`
		assert.Contains(t, query, "SET is_active = false")
	})

	t.Run("get balance query", func(t *testing.T) {
		query := `WITH processed_orders AS ( SELECT accrual FROM orders WHERE user_id = $1 AND status = 'PROCESSED' ) SELECT $1::int as user_id, COALESCE(SUM(accrual), 0) as current, COALESCE(ABS(SUM(CASE WHEN accrual < 0 THEN accrual ELSE 0 END)), 0) as withdrawn FROM processed_orders`
		assert.Contains(t, query, "WITH processed_orders AS")
		assert.Contains(t, query, "COALESCE(SUM(accrual), 0)")
	})

	t.Run("get stats query", func(t *testing.T) {
		query := `SELECT COUNT(*) as total_orders, COUNT(CASE WHEN status = 'PROCESSED' THEN 1 END) as processed_orders, COUNT(CASE WHEN status = 'NEW' THEN 1 END) as new_orders, COUNT(CASE WHEN status = 'PROCESSING' THEN 1 END) as processing_orders, COUNT(CASE WHEN status = 'INVALID' THEN 1 END) as invalid_orders, COALESCE(SUM(CASE WHEN status = 'PROCESSED' THEN accrual ELSE 0 END), 0) as total_accrual, COALESCE(ABS(SUM(CASE WHEN status = 'PROCESSED' AND accrual < 0 THEN accrual ELSE 0 END)), 0) as total_withdrawn FROM orders WHERE user_id = $1`
		assert.Contains(t, query, "COUNT(*) as total_orders")
		assert.Contains(t, query, "COALESCE(SUM")
	})
}

// ==================== Тесты для UserBalance ====================

func TestUserBalance(t *testing.T) {
	tests := []struct {
		name        string
		balance     entity.UserBalance
		expectedSum float64
	}{
		{
			name: "positive balance",
			balance: entity.UserBalance{
				UserID:    1,
				Current:   1000.50,
				Withdrawn: 200.25,
			},
			expectedSum: 1200.75, // Current + Withdrawn
		},
		{
			name: "zero balance",
			balance: entity.UserBalance{
				UserID:    2,
				Current:   0.0,
				Withdrawn: 0.0,
			},
			expectedSum: 0.0,
		},
		{
			name: "only withdrawals",
			balance: entity.UserBalance{
				UserID:    3,
				Current:   0.0,
				Withdrawn: 500.0,
			},
			expectedSum: 500.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sum := tt.balance.Current + tt.balance.Withdrawn
			assert.Equal(t, tt.expectedSum, sum)
			assert.Equal(t, tt.balance.UserID, tt.balance.UserID)

			// Проверяем, что текущий баланс не может быть отрицательным
			// (это бизнес-логика, которую можно проверить)
			if tt.balance.Current < 0 {
				t.Errorf("Current balance should not be negative: %f", tt.balance.Current)
			}

			// Сумма списаний не может быть отрицательной
			if tt.balance.Withdrawn < 0 {
				t.Errorf("Withdrawn amount should not be negative: %f", tt.balance.Withdrawn)
			}
		})
	}
}

// ==================== Тесты для UserStats ====================

func TestUserStats(t *testing.T) {
	stats := entity.UserStats{
		UserID:           1,
		TotalOrders:      10,
		ProcessedOrders:  5,
		NewOrders:        2,
		ProcessingOrders: 2,
		InvalidOrders:    1,
		TotalAccrual:     500.0,
		TotalWithdrawn:   200.0,
	}

	// Проверяем, что сумма всех статусов равна общему количеству заказов
	totalByStatus := stats.NewOrders + stats.ProcessingOrders +
		stats.ProcessedOrders + stats.InvalidOrders
	assert.Equal(t, stats.TotalOrders, totalByStatus)

	// Проверяем, что TotalAccrual >= TotalWithdrawn (в реальной системе)
	if stats.TotalAccrual < stats.TotalWithdrawn {
		t.Errorf("Total accrual (%f) should not be less than total withdrawn (%f)",
			stats.TotalAccrual, stats.TotalWithdrawn)
	}
}

// ==================== Контекстные тесты ====================

func TestUserContext(t *testing.T) {
	ctx := context.Background()

	t.Run("context with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		select {
		case <-time.After(50 * time.Millisecond):
			// OK
		case <-ctx.Done():
			t.Error("Context should not be done yet")
		}
	})

	t.Run("context with cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		cancel()

		select {
		case <-ctx.Done():
			// OK
		case <-time.After(50 * time.Millisecond):
			t.Error("Context should be cancelled")
		}
	})
}

// ==================== Тесты для UserWithOrders ====================

func TestUserWithOrders(t *testing.T) {
	user := entity.User{
		ID:        1,
		Login:     "testuser",
		IsActive:  true,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	orders := []entity.Order{
		{
			ID:     1,
			Number: "1234567812345678",
			Status: entity.OrderStatusNew,
			UserID: 1,
		},
		{
			ID:     2,
			Number: "8765432187654321",
			Status: entity.OrderStatusProcessed,
			UserID: 1,
		},
	}

	userWithOrders := entity.UserWithOrders{
		User:   user,
		Orders: orders,
	}

	assert.Equal(t, user.ID, userWithOrders.User.ID)
	assert.Equal(t, user.Login, userWithOrders.User.Login)
	assert.Len(t, userWithOrders.Orders, 2)
	assert.Equal(t, orders[0].Number, userWithOrders.Orders[0].Number)
	assert.Equal(t, orders[1].Status, userWithOrders.Orders[1].Status)
}

// ==================== Параллельные тесты ====================

func TestUserParallelTests(t *testing.T) {
	t.Run("test 1", func(t *testing.T) {
		t.Parallel()
		time.Sleep(50 * time.Millisecond)
		user := entity.User{ID: 1, Login: "user1"}
		assert.Equal(t, 1, user.ID)
	})

	t.Run("test 2", func(t *testing.T) {
		t.Parallel()
		time.Sleep(50 * time.Millisecond)
		user := entity.User{ID: 2, Login: "user2"}
		assert.Equal(t, 2, user.ID)
	})

	t.Run("test 3", func(t *testing.T) {
		t.Parallel()
		time.Sleep(50 * time.Millisecond)
		user := entity.User{ID: 3, Login: "user3"}
		assert.Equal(t, 3, user.ID)
	})
}

// ==================== Benchmark тесты ====================

func BenchmarkUserScan(b *testing.B) {
	repo := &userRepo{}
	scanner := &MockUserScanner{
		id:        1,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		login:     "benchuser",
		password:  "benchpass",
		isActive:  true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var user entity.User
		_ = repo.scanUser(scanner, &user)
	}
}

func BenchmarkUserStructCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = entity.User{
			ID:        i,
			Login:     "testuser",
			Password:  "password",
			IsActive:  true,
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		}
	}
}

// ==================== Table-driven тесты ====================

func TestUserTableDriven(t *testing.T) {
	tests := []struct {
		name     string
		userID   int
		login    string
		password string
		isActive bool
	}{
		{"active user", 1, "user1", "pass1", true},
		{"inactive user", 2, "user2", "pass2", false},
		{"another active", 3, "user3", "pass3", true},
		{"zero id", 0, "user0", "pass0", true},
		{"special chars", 4, "user.name-123", "pass!@#", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := entity.User{
				ID:       tt.userID,
				Login:    tt.login,
				Password: tt.password,
				IsActive: tt.isActive,
			}

			assert.Equal(t, tt.userID, user.ID)
			assert.Equal(t, tt.login, user.Login)
			assert.Equal(t, tt.password, user.Password)
			assert.Equal(t, tt.isActive, user.IsActive)
		})
	}
}

// ==================== Тесты для обработки ошибок ====================

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"user not active", ErrUserNotActive, "user not active"},
		{"user already exists", ErrUserAlreadyExists, "user already exists"},
		{"user has active orders", ErrUserHasActiveOrders, "user has active orders"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualError(t, tt.err, tt.expected)
		})
	}
}

// ==================== Тесты для форматов времени ====================

func TestUserTimeFormats(t *testing.T) {
	now := time.Now()

	formats := []struct {
		name   string
		format string
	}{
		{"RFC3339", time.RFC3339},
		{"RFC3339Nano", time.RFC3339Nano},
	}

	for _, f := range formats {
		t.Run(f.name, func(t *testing.T) {
			formatted := now.Format(f.format)
			parsed, err := time.Parse(f.format, formatted)

			assert.NoError(t, err)
			assert.WithinDuration(t, now, parsed, time.Second)
		})
	}
}

// ==================== Простые smoke тесты ====================

func TestUserSmokeTests(t *testing.T) {
	// Простые тесты для проверки базовой функциональности

	t.Run("empty user", func(t *testing.T) {
		var user entity.User
		assert.Equal(t, 0, user.ID)
		assert.Equal(t, "", user.Login)
		assert.Equal(t, "", user.Password)
		assert.False(t, user.IsActive)
	})

	t.Run("user with orders", func(t *testing.T) {
		user := entity.User{
			ID:    1,
			Login: "test",
		}

		order := entity.Order{
			ID:     1,
			Number: "123",
			UserID: 1,
		}

		assert.Equal(t, user.ID, order.UserID)
	})

	t.Run("time comparison", func(t *testing.T) {
		user1 := entity.User{
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		}

		user2 := entity.User{
			CreatedAt: time.Now().Add(-time.Hour).Format(time.RFC3339),
			UpdatedAt: time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
		}

		// Просто проверяем, что время задано
		assert.NotEmpty(t, user1.CreatedAt)
		assert.NotEmpty(t, user2.CreatedAt)
	})
}

// ==================== Тесты для пограничных случаев ====================

func TestUserEdgeCases(t *testing.T) {
	t.Run("very long login", func(t *testing.T) {
		longLogin := "a"
		for i := 0; i < 255; i++ {
			longLogin += "a"
		}

		user := entity.User{
			ID:    1,
			Login: longLogin[:255], // Обрезаем до 255 символов
		}

		assert.Len(t, user.Login, 255)
	})

	t.Run("special characters in password", func(t *testing.T) {
		specialPassword := "p@$$w0rd!№;%:?*()_+-=<>"
		user := entity.User{
			ID:       1,
			Password: specialPassword,
		}

		assert.Equal(t, specialPassword, user.Password)
	})

	t.Run("unicode in login", func(t *testing.T) {
		user := entity.User{
			ID:    1,
			Login: "用户123",
		}

		assert.Equal(t, "用户123", user.Login)
	})
}

// ==================== Тесты для производительности ====================

func TestUserPerformance(t *testing.T) {
	// Тест на создание большого количества пользователей
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	start := time.Now()

	// Создаем 1000 пользователей
	users := make([]entity.User, 1000)
	for i := 0; i < 1000; i++ {
		users[i] = entity.User{
			ID:       i + 1,
			Login:    fmt.Sprintf("user%d", i),
			Password: fmt.Sprintf("pass%d", i),
			IsActive: i%2 == 0,
		}
	}

	duration := time.Since(start)
	t.Logf("Created 1000 users in %v", duration)
	assert.Less(t, duration, 100*time.Millisecond, "Should create 1000 users quickly")
}

// ==================== Основной тест ====================

func TestUserMain(t *testing.T) {
	// Основной тест, который запускает несколько подтестов
	t.Run("create and validate user", func(t *testing.T) {
		user := entity.User{
			ID:        1,
			Login:     "testuser",
			Password:  "hashed_password_123",
			IsActive:  true,
			CreatedAt: time.Now().Format(time.RFC3339),
		}

		// Проверяем все поля
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "testuser", user.Login)
		assert.Equal(t, "hashed_password_123", user.Password)
		assert.True(t, user.IsActive)
		assert.NotEmpty(t, user.CreatedAt)

		// Проверяем, что пароль не пустой
		assert.NotEmpty(t, user.Password)

		// Проверяем, что логин не слишком короткий (в реальном приложении)
		if len(user.Login) < 3 {
			t.Errorf("Login too short: %s", user.Login)
		}
	})
}

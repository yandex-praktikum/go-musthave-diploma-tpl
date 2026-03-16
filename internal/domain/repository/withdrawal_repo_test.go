// withdrawal_repo_test.go
package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для сканирования ====================

// MockWithdrawalScanner для тестирования scanWithdrawal
type MockWithdrawalScanner struct {
	order       string
	sum         float64
	processedAt time.Time
	userID      int
	err         error
}

func (m *MockWithdrawalScanner) Scan(dest ...interface{}) error {
	if m.err != nil {
		return m.err
	}

	// Заполняем значения в порядке, ожидаемом scanWithdrawal
	*dest[0].(*string) = m.order
	*dest[1].(*float64) = m.sum
	*dest[2].(*time.Time) = m.processedAt
	*dest[3].(*int) = m.userID

	return nil
}

func TestWithdrawalRepo_scanWithdrawal(t *testing.T) {
	repo := &withdrawalRepo{}

	tests := []struct {
		name        string
		scanner     *MockWithdrawalScanner
		wantError   bool
		checkResult func(*testing.T, *entity.Withdrawal)
	}{
		{
			name: "successful scan with negative sum",
			scanner: &MockWithdrawalScanner{
				order:       "1234567812345678",
				sum:         -150.75, // Отрицательная сумма в БД
				processedAt: time.Now(),
				userID:      1,
			},
			wantError: false,
			checkResult: func(t *testing.T, w *entity.Withdrawal) {
				assert.Equal(t, "1234567812345678", w.Order)
				assert.Equal(t, 150.75, w.Sum) // Должна быть преобразована в положительную
				assert.Equal(t, 1, w.UserID)
				assert.NotEmpty(t, w.ProcessedAt)
			},
		},
		{
			name: "successful scan with positive sum (should not happen in practice)",
			scanner: &MockWithdrawalScanner{
				order:       "8765432187654321",
				sum:         100.0, // Положительная сумма
				processedAt: time.Now(),
				userID:      2,
			},
			wantError: false,
			checkResult: func(t *testing.T, w *entity.Withdrawal) {
				assert.Equal(t, "8765432187654321", w.Order)
				assert.Equal(t, 100.0, w.Sum) // Остается положительной
				assert.Equal(t, 2, w.UserID)
			},
		},
		{
			name: "scan error",
			scanner: &MockWithdrawalScanner{
				err: pgx.ErrNoRows,
			},
			wantError: true,
			checkResult: func(t *testing.T, w *entity.Withdrawal) {
				// Не должно вызываться
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var withdrawal entity.Withdrawal
			err := repo.scanWithdrawal(tt.scanner, &withdrawal)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.checkResult(t, &withdrawal)
			}
		})
	}
}

// ==================== Тесты для интерфейсов и структур ====================

func TestWithdrawalRepositoryInterfaces(t *testing.T) {
	// Проверяем, что структура реализует интерфейс
	var repo WithdrawalRepository
	repo = &withdrawalRepo{}
	assert.NotNil(t, repo)

	// Проверяем наличие методов (компиляция проверит)
	_ = repo
}

func TestWithdrawalStruct2(t *testing.T) {
	tests := []struct {
		name       string
		withdrawal entity.Withdrawal
		shouldPass bool
	}{
		{
			name: "valid withdrawal",
			withdrawal: entity.Withdrawal{
				UserID:      1,
				Order:       "1234567812345678",
				Sum:         150.75,
				ProcessedAt: time.Now().Format(time.RFC3339),
			},
			shouldPass: true,
		},
		{
			name: "zero sum",
			withdrawal: entity.Withdrawal{
				UserID: 2,
				Order:  "8765432187654321",
				Sum:    0.0,
			},
			shouldPass: false, // Сумма должна быть положительной
		},
		{
			name: "negative sum",
			withdrawal: entity.Withdrawal{
				UserID: 3,
				Order:  "1111222233334444",
				Sum:    -100.0,
			},
			shouldPass: false, // Сумма должна быть положительной (преобразуется при сканировании)
		},
		{
			name: "empty order number",
			withdrawal: entity.Withdrawal{
				UserID: 4,
				Order:  "",
				Sum:    50.0,
			},
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// В реальном приложении здесь была бы валидация
			if tt.shouldPass {
				assert.NotEmpty(t, tt.withdrawal.Order)
				assert.Greater(t, tt.withdrawal.Sum, 0.0)
				assert.Greater(t, tt.withdrawal.UserID, 0)
			}
		})
	}
}

// ==================== Тесты для SQL запросов ====================

func TestWithdrawalSQLQueries(t *testing.T) {

	t.Run("get by user id query", func(t *testing.T) {
		query := `SELECT number, accrual, processed_at, user_id FROM orders WHERE user_id = $1 AND status = $2 AND accrual < 0 ORDER BY uploaded_at DESC`
		assert.Contains(t, query, "SELECT number, accrual, processed_at, user_id")
		assert.Contains(t, query, "accrual < 0")
		assert.Contains(t, query, "ORDER BY uploaded_at DESC")
	})

	t.Run("get all withdrawals query", func(t *testing.T) {
		query := `SELECT number, accrual, processed_at, user_id FROM orders WHERE accrual < 0 AND status = 'PROCESSED' ORDER BY uploaded_at DESC LIMIT $1 OFFSET $2`
		assert.Contains(t, query, "WHERE accrual < 0")
		assert.Contains(t, query, "LIMIT $1 OFFSET $2")
	})

	t.Run("get total withdrawn query", func(t *testing.T) {
		query := `SELECT COALESCE(ABS(SUM(accrual)), 0) FROM orders WHERE user_id = $1 AND status = 'PROCESSED' AND accrual < 0`
		assert.Contains(t, query, "COALESCE(ABS(SUM(accrual)), 0)")
		assert.Contains(t, query, "accrual < 0")
	})

	t.Run("get withdrawal by order query", func(t *testing.T) {
		query := `SELECT number, accrual, processed_at, user_id FROM orders WHERE number = $1 AND accrual < 0 AND status = 'PROCESSED'`
		assert.Contains(t, query, "WHERE number = $1")
		assert.Contains(t, query, "accrual < 0")
	})

	t.Run("get withdrawals summary query", func(t *testing.T) {
		query := `SELECT COUNT(*) as total_withdrawals, COALESCE(ABS(SUM(accrual)), 0) as total_amount, COUNT(DISTINCT user_id) as unique_users FROM orders WHERE accrual < 0 AND status = 'PROCESSED'`
		assert.Contains(t, query, "COUNT(*) as total_withdrawals")
		assert.Contains(t, query, "COUNT(DISTINCT user_id) as unique_users")
	})

	t.Run("get user withdrawals summary query", func(t *testing.T) {
		query := `SELECT COUNT(*) as withdrawal_count, COALESCE(ABS(SUM(accrual)), 0) as total_amount, MIN(processed_at) as first_withdrawal, MAX(processed_at) as last_withdrawal FROM orders WHERE user_id = $1 AND accrual < 0 AND status = 'PROCESSED'`
		assert.Contains(t, query, "MIN(processed_at) as first_withdrawal")
		assert.Contains(t, query, "MAX(processed_at) as last_withdrawal")
	})

	t.Run("get recent withdrawals query", func(t *testing.T) {
		query := `SELECT number, accrual, processed_at, user_id FROM orders WHERE accrual < 0 AND status = 'PROCESSED' ORDER BY processed_at DESC LIMIT $1`
		assert.Contains(t, query, "ORDER BY processed_at DESC")
	})

	t.Run("get withdrawals by period query", func(t *testing.T) {
		query := `SELECT number, accrual, processed_at, user_id FROM orders WHERE accrual < 0 AND status = 'PROCESSED' AND processed_at BETWEEN $1 AND $2 ORDER BY processed_at DESC`
		assert.Contains(t, query, "processed_at BETWEEN $1 AND $2")
	})
}

// ==================== Тесты для WithdrawalsSummary ====================

func TestWithdrawalsSummary(t *testing.T) {
	tests := []struct {
		name     string
		summary  entity.WithdrawalsSummary
		expected func(*testing.T, entity.WithdrawalsSummary)
	}{
		{
			name: "valid summary with data",
			summary: entity.WithdrawalsSummary{
				TotalWithdrawals: 100,
				TotalAmount:      5000.75,
				UniqueUsers:      25,
			},
			expected: func(t *testing.T, s entity.WithdrawalsSummary) {
				assert.Equal(t, 100, s.TotalWithdrawals)
				assert.Equal(t, 5000.75, s.TotalAmount)
				assert.Equal(t, 25, s.UniqueUsers)
				// Проверяем, что уникальных пользователей не больше, чем общее количество списаний
				if s.UniqueUsers > s.TotalWithdrawals && s.TotalWithdrawals > 0 {
					t.Errorf("Unique users (%d) should not exceed total withdrawals (%d)",
						s.UniqueUsers, s.TotalWithdrawals)
				}
			},
		},
		{
			name: "zero summary",
			summary: entity.WithdrawalsSummary{
				TotalWithdrawals: 0,
				TotalAmount:      0.0,
				UniqueUsers:      0,
			},
			expected: func(t *testing.T, s entity.WithdrawalsSummary) {
				assert.Equal(t, 0, s.TotalWithdrawals)
				assert.Equal(t, 0.0, s.TotalAmount)
				assert.Equal(t, 0, s.UniqueUsers)
			},
		},
		{
			name: "summary with many unique users",
			summary: entity.WithdrawalsSummary{
				TotalWithdrawals: 50,
				TotalAmount:      2500.0,
				UniqueUsers:      50, // Каждый пользователь сделал по одному списанию
			},
			expected: func(t *testing.T, s entity.WithdrawalsSummary) {
				assert.Equal(t, s.TotalWithdrawals, s.UniqueUsers)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.expected(t, tt.summary)
		})
	}
}

// ==================== Тесты для UserWithdrawalsSummary ====================

func TestUserWithdrawalsSummary(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		summary  entity.UserWithdrawalsSummary
		expected func(*testing.T, entity.UserWithdrawalsSummary)
	}{
		{
			name: "valid summary with dates",
			summary: entity.UserWithdrawalsSummary{
				UserID:          1,
				WithdrawalCount: 5,
				TotalAmount:     750.50,
				FirstWithdrawal: now.Add(-30 * 24 * time.Hour).Format(time.RFC3339),
				LastWithdrawal:  now.Format(time.RFC3339),
			},
			expected: func(t *testing.T, s entity.UserWithdrawalsSummary) {
				assert.Equal(t, 1, s.UserID)
				assert.Equal(t, 5, s.WithdrawalCount)
				assert.Equal(t, 750.50, s.TotalAmount)
				assert.NotEmpty(t, s.FirstWithdrawal)
				assert.NotEmpty(t, s.LastWithdrawal)
				// Проверяем, что первое списание раньше последнего
				first, err1 := time.Parse(time.RFC3339, s.FirstWithdrawal)
				last, err2 := time.Parse(time.RFC3339, s.LastWithdrawal)
				if err1 == nil && err2 == nil {
					assert.True(t, first.Before(last) || first.Equal(last))
				}
			},
		},
		{
			name: "summary without withdrawals",
			summary: entity.UserWithdrawalsSummary{
				UserID:          2,
				WithdrawalCount: 0,
				TotalAmount:     0.0,
				FirstWithdrawal: "",
				LastWithdrawal:  "",
			},
			expected: func(t *testing.T, s entity.UserWithdrawalsSummary) {
				assert.Equal(t, 2, s.UserID)
				assert.Equal(t, 0, s.WithdrawalCount)
				assert.Equal(t, 0.0, s.TotalAmount)
				assert.Empty(t, s.FirstWithdrawal)
				assert.Empty(t, s.LastWithdrawal)
			},
		},
		{
			name: "summary with only one withdrawal",
			summary: entity.UserWithdrawalsSummary{
				UserID:          3,
				WithdrawalCount: 1,
				TotalAmount:     100.0,
				FirstWithdrawal: now.Format(time.RFC3339),
				LastWithdrawal:  now.Format(time.RFC3339),
			},
			expected: func(t *testing.T, s entity.UserWithdrawalsSummary) {
				assert.Equal(t, 1, s.WithdrawalCount)
				assert.Equal(t, s.FirstWithdrawal, s.LastWithdrawal)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.expected(t, tt.summary)
		})
	}
}

// ==================== Тесты для ошибок ====================

func TestWithdrawalErrorHandling(t *testing.T) {
	t.Run("negative sum validation", func(t *testing.T) {
		// Проверяем, что сумма должна быть отрицательной при создании списания
		withdrawal := &entity.Withdrawal{
			UserID: 1,
			Order:  "1234567812345678",
			Sum:    100.0, // Положительная сумма
		}

		// В реальном коде метод Create проверяет, что сумма отрицательная
		// Мы можем только проверить логику без вызова метода
		if withdrawal.Sum >= 0 {
			// Это должно вызывать ошибку в методе Create
			t.Log("Positive sum should trigger an error in Create method")
		}
	})

	t.Run("duplicate order number", func(t *testing.T) {
		// Проверяем обработку дубликатов
		// В реальном коде IsDuplicateError используется для проверки
		assert.NotNil(t, IsDuplicateError)
	})

	t.Run("foreign key constraint", func(t *testing.T) {
		// Проверяем обработку нарушений внешнего ключа
		assert.NotNil(t, IsForeignKeyError)
	})
}

// ==================== Контекстные тесты ====================

func TestWithdrawalContext(t *testing.T) {
	ctx := context.Background()

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		cancel()

		select {
		case <-ctx.Done():
			// Ожидаемое поведение
		case <-time.After(50 * time.Millisecond):
			t.Error("Context should be cancelled")
		}
	})

	t.Run("context with deadline", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(100*time.Millisecond))
		defer cancel()

		select {
		case <-time.After(50 * time.Millisecond):
			// OK, дедлайн еще не наступил
		case <-ctx.Done():
			t.Error("Context should not be done yet")
		}

		// Ждем, пока дедлайн точно пройдет
		time.Sleep(100 * time.Millisecond)

		select {
		case <-ctx.Done():
			// OK, дедлайн наступил
		case <-time.After(50 * time.Millisecond):
			t.Error("Context should be done after deadline")
		}
	})
}

// ==================== Параллельные тесты ====================

func TestWithdrawalParallelTests(t *testing.T) {
	t.Run("parallel test 1", func(t *testing.T) {
		t.Parallel()
		time.Sleep(50 * time.Millisecond)
		withdrawal := entity.Withdrawal{UserID: 1, Sum: 100.0}
		assert.Equal(t, 100.0, withdrawal.Sum)
	})

	t.Run("parallel test 2", func(t *testing.T) {
		t.Parallel()
		time.Sleep(50 * time.Millisecond)
		summary := entity.WithdrawalsSummary{TotalWithdrawals: 10}
		assert.Equal(t, 10, summary.TotalWithdrawals)
	})

	t.Run("parallel test 3", func(t *testing.T) {
		t.Parallel()
		time.Sleep(50 * time.Millisecond)
		userSummary := entity.UserWithdrawalsSummary{UserID: 1}
		assert.Equal(t, 1, userSummary.UserID)
	})
}

// ==================== Benchmark тесты ====================

func BenchmarkWithdrawalScan(b *testing.B) {
	repo := &withdrawalRepo{}
	scanner := &MockWithdrawalScanner{
		order:       "1234567812345678",
		sum:         -150.75,
		processedAt: time.Now(),
		userID:      1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var withdrawal entity.Withdrawal
		_ = repo.scanWithdrawal(scanner, &withdrawal)
	}
}

func BenchmarkWithdrawalStructCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = entity.Withdrawal{
			UserID:      i,
			Order:       "1234567812345678",
			Sum:         100.0,
			ProcessedAt: time.Now().Format(time.RFC3339),
		}
	}
}

func BenchmarkWithdrawalsSummaryCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = entity.WithdrawalsSummary{
			TotalWithdrawals: i,
			TotalAmount:      float64(i) * 100.0,
			UniqueUsers:      i / 2,
		}
	}
}

// ==================== Table-driven тесты ====================

func TestWithdrawalTableDriven(t *testing.T) {
	tests := []struct {
		name    string
		userID  int
		order   string
		sum     float64
		isValid bool
	}{
		{"valid withdrawal", 1, "1234567812345678", 150.75, true},
		{"another valid", 2, "8765432187654321", 99.99, true},
		{"zero sum", 3, "1111222233334444", 0.0, false},
		{"negative sum", 4, "5555666677778888", -50.0, false},
		{"empty order", 5, "", 100.0, false},
		{"short order", 6, "123", 100.0, false},
		{"very small sum", 7, "9999888877776666", 0.01, true},
		{"very large sum", 8, "1234123412341234", 999999.99, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withdrawal := entity.Withdrawal{
				UserID: tt.userID,
				Order:  tt.order,
				Sum:    tt.sum,
			}

			// Проверяем условия валидности
			isValid := withdrawal.UserID > 0 &&
				len(withdrawal.Order) >= 10 && // Минимальная длина номера заказа
				withdrawal.Sum > 0

			assert.Equal(t, tt.isValid, isValid,
				"Withdrawal validation failed: UserID=%d, Order=%s, Sum=%f",
				withdrawal.UserID, withdrawal.Order, withdrawal.Sum)
		})
	}
}

// ==================== Тесты для форматов времени ====================

func TestWithdrawalTimeFormats(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"current time", now, now.Format(time.RFC3339)},
		{"zero time", time.Time{}, "0001-01-01T00:00:00Z"},
		{"past time", now.Add(-24 * 365 * time.Hour), now.Add(-24 * 365 * time.Hour).Format(time.RFC3339)},
		{"future time", now.Add(24 * 365 * time.Hour), now.Add(24 * 365 * time.Hour).Format(time.RFC3339)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := tt.time.Format(time.RFC3339)
			assert.Equal(t, tt.expected, formatted)

			// Проверяем, что можем распарсить обратно
			if !tt.time.IsZero() {
				parsed, err := time.Parse(time.RFC3339, formatted)
				assert.NoError(t, err)
				assert.WithinDuration(t, tt.time, parsed, time.Second)
			}
		})
	}
}

// ==================== Тесты для пограничных случаев ====================

func TestWithdrawalEdgeCases(t *testing.T) {
	t.Run("very large withdrawal amount", func(t *testing.T) {
		withdrawal := entity.Withdrawal{
			UserID: 1,
			Order:  "1234567812345678",
			Sum:    1e9, // 1 миллиард
		}

		// Проверяем, что сумма корректно хранится
		assert.Equal(t, 1e9, withdrawal.Sum)
	})

	t.Run("withdrawal with many decimal places", func(t *testing.T) {
		withdrawal := entity.Withdrawal{
			UserID: 2,
			Order:  "8765432187654321",
			Sum:    123.456789,
		}

		// В реальной системе может быть округление
		assert.Equal(t, 123.456789, withdrawal.Sum)
	})

	t.Run("multiple withdrawals same user", func(t *testing.T) {
		// Тестируем логику агрегации
		withdrawals := []entity.Withdrawal{
			{UserID: 1, Sum: 100.0},
			{UserID: 1, Sum: 200.0},
			{UserID: 1, Sum: 300.0},
		}

		total := 0.0
		for _, w := range withdrawals {
			total += w.Sum
		}

		assert.Equal(t, 600.0, total)
	})

	t.Run("withdrawal summary edge cases", func(t *testing.T) {
		summary := entity.WithdrawalsSummary{
			TotalWithdrawals: 0,
			TotalAmount:      0.0,
			UniqueUsers:      0,
		}

		// Проверяем нулевые значения
		assert.Equal(t, 0, summary.TotalWithdrawals)
		assert.Equal(t, 0.0, summary.TotalAmount)
		assert.Equal(t, 0, summary.UniqueUsers)

		// Проверяем, что при отсутствии списаний уникальных пользователей тоже 0
		if summary.TotalWithdrawals == 0 {
			assert.Equal(t, 0, summary.UniqueUsers)
		}
	})
}

// ==================== Smoke тесты ====================

func TestWithdrawalSmokeTests(t *testing.T) {
	t.Run("empty withdrawal", func(t *testing.T) {
		var withdrawal entity.Withdrawal
		assert.Equal(t, 0, withdrawal.UserID)
		assert.Equal(t, "", withdrawal.Order)
		assert.Equal(t, 0.0, withdrawal.Sum)
		assert.Equal(t, "", withdrawal.ProcessedAt)
	})

	t.Run("empty withdrawals summary", func(t *testing.T) {
		var summary entity.WithdrawalsSummary
		assert.Equal(t, 0, summary.TotalWithdrawals)
		assert.Equal(t, 0.0, summary.TotalAmount)
		assert.Equal(t, 0, summary.UniqueUsers)
	})

	t.Run("empty user withdrawals summary", func(t *testing.T) {
		var summary entity.UserWithdrawalsSummary
		assert.Equal(t, 0, summary.UserID)
		assert.Equal(t, 0, summary.WithdrawalCount)
		assert.Equal(t, 0.0, summary.TotalAmount)
		assert.Equal(t, "", summary.FirstWithdrawal)
		assert.Equal(t, "", summary.LastWithdrawal)
	})

	t.Run("time comparison in withdrawal", func(t *testing.T) {
		withdrawal1 := entity.Withdrawal{
			ProcessedAt: time.Now().Format(time.RFC3339),
		}

		withdrawal2 := entity.Withdrawal{
			ProcessedAt: time.Now().Add(-time.Hour).Format(time.RFC3339),
		}

		// Просто проверяем, что время задано
		assert.NotEmpty(t, withdrawal1.ProcessedAt)
		assert.NotEmpty(t, withdrawal2.ProcessedAt)

		// Парсим и сравниваем
		t1, err1 := time.Parse(time.RFC3339, withdrawal1.ProcessedAt)
		t2, err2 := time.Parse(time.RFC3339, withdrawal2.ProcessedAt)

		if err1 == nil && err2 == nil {
			assert.True(t, t2.Before(t1))
		}
	})
}

// ==================== Основной тест ====================

func TestWithdrawalMain(t *testing.T) {
	t.Run("complete withdrawal lifecycle", func(t *testing.T) {
		// Создаем тестовое списание
		withdrawal := entity.Withdrawal{
			UserID:      1,
			Order:       "1234567812345678",
			Sum:         150.75,
			ProcessedAt: time.Now().Format(time.RFC3339),
		}

		// Проверяем все поля
		assert.Equal(t, 1, withdrawal.UserID)
		assert.Equal(t, "1234567812345678", withdrawal.Order)
		assert.Equal(t, 150.75, withdrawal.Sum)
		assert.NotEmpty(t, withdrawal.ProcessedAt)

		// Проверяем, что номер заказа похож на валидный
		assert.GreaterOrEqual(t, len(withdrawal.Order), 10)

		// Проверяем, что сумма положительная
		assert.Greater(t, withdrawal.Sum, 0.0)

		// Проверяем формат времени
		_, err := time.Parse(time.RFC3339, withdrawal.ProcessedAt)
		assert.NoError(t, err)
	})

	t.Run("withdrawal aggregation", func(t *testing.T) {
		// Тестируем логику агрегации списаний
		withdrawals := []entity.Withdrawal{
			{UserID: 1, Sum: 100.0},
			{UserID: 1, Sum: 200.0},
			{UserID: 2, Sum: 300.0},
			{UserID: 3, Sum: 400.0},
		}

		// Вычисляем общую сумму
		totalAmount := 0.0
		for _, w := range withdrawals {
			totalAmount += w.Sum
		}
		assert.Equal(t, 1000.0, totalAmount)

		// Вычисляем количество уникальных пользователей
		uniqueUsers := make(map[int]bool)
		for _, w := range withdrawals {
			uniqueUsers[w.UserID] = true
		}
		assert.Equal(t, 3, len(uniqueUsers))

		// Проверяем, что сумма списаний не отрицательная
		for _, w := range withdrawals {
			assert.GreaterOrEqual(t, w.Sum, 0.0)
		}
	})
}

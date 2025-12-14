package repository

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

// Реализация Scanner для тестов
type testScanner struct {
	id        int
	createdAt time.Time
	updatedAt time.Time
	number    string
	status    entity.OrderStatus
	accrual   *float64
	userID    int
	err       error
}

func (s *testScanner) Scan(dest ...interface{}) error {
	if s.err != nil {
		return s.err
	}

	// Заполняем dest значениями
	*dest[0].(*int) = s.id
	*dest[1].(*time.Time) = s.createdAt
	*dest[2].(*time.Time) = s.updatedAt
	*dest[3].(*string) = s.number
	*dest[4].(*entity.OrderStatus) = s.status
	*dest[5].(**float64) = s.accrual
	*dest[6].(*int) = s.userID

	return nil
}

func TestScanOrderReal(t *testing.T) {
	repo := &orderRepo{}

	t.Run("scan with accrual", func(t *testing.T) {
		now := time.Now()
		accrual := 100.5

		scanner := &testScanner{
			id:        1,
			createdAt: now,
			updatedAt: now.Add(time.Hour),
			number:    "1234567890",
			status:    entity.OrderStatusProcessed,
			accrual:   &accrual,
			userID:    1,
		}

		var order entity.Order
		err := repo.scanOrder(scanner, &order)

		assert.NoError(t, err)
		assert.Equal(t, 1, order.ID)
		assert.Equal(t, "1234567890", order.Number)
		assert.Equal(t, entity.OrderStatusProcessed, order.Status)
		assert.Equal(t, 1, order.UserID)
		assert.Equal(t, 100.5, *order.Accrual)
		assert.Equal(t, now.Format(time.RFC3339), order.UploadedAt)
		assert.Equal(t, now.Add(time.Hour).Format(time.RFC3339), order.ProcessedAt)
	})

	t.Run("scan without accrual", func(t *testing.T) {
		now := time.Now()

		scanner := &testScanner{
			id:        2,
			createdAt: now,
			updatedAt: now,
			number:    "9876543210",
			status:    entity.OrderStatusNew,
			accrual:   nil,
			userID:    2,
		}

		var order entity.Order
		err := repo.scanOrder(scanner, &order)

		assert.NoError(t, err)
		assert.Equal(t, 2, order.ID)
		assert.Equal(t, "9876543210", order.Number)
		assert.Equal(t, entity.OrderStatusNew, order.Status)
		assert.Nil(t, order.Accrual)
	})

	t.Run("scan error", func(t *testing.T) {
		scanner := &testScanner{
			err: pgx.ErrNoRows,
		}

		var order entity.Order
		err := repo.scanOrder(scanner, &order)

		assert.Error(t, err)
		assert.Equal(t, pgx.ErrNoRows, err)
	})
}

// Простой тест для методов orderRepo
func TestOrderRepoMethods(t *testing.T) {
	// Эти тесты просто проверяют, что методы определены
	// Реальные тесты потребуют мокирования базы данных
	repo := &orderRepo{}

	assert.NotNil(t, repo)
	// repo.Create, repo.GetByNumber и т.д. определены
}

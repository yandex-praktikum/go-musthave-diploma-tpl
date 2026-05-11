//go:build e2e

package postgresql_test

import (
	"testing"

	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/repositories/postgresql"
	"github.com/KalessinD/gophermart/tests/e2e/fixtures"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type OrderE2ETestSuite struct {
	fixtures.PostgresSuite
	storage postgresql.SQLStorageInterface
}

func (s *OrderE2ETestSuite) SetupSuite() {
	s.PostgresSuite.SetupSuite()
	s.storage = postgresql.NewSQLStorage(s.DB)
}

func (s *OrderE2ETestSuite) SetupTest() {
	// Очищаем и orders, и users, так как заказ зависит от пользователя
	s.PostgresSuite.SetupTest([]string{"gophermart.orders", "gophermart.users"})
}

func TestOrderE2ETestSuite(t *testing.T) {
	suite.Run(t, new(OrderE2ETestSuite))
}

// --- Тесты ---

func (s *OrderE2ETestSuite) TestAddAndGetOrder() {
	user := &models.User{
		ID:    uuid.New().String(),
		Login: "order_user_1",
		Hash:  "hash",
	}
	err := s.storage.AddUser(s.Ctx, user)
	require.NoError(s.T(), err, "Failed to create user for order")

	order := &models.Order{
		ID:     "12345678903",
		UserID: user.ID,
		Status: models.OrderNewStatus,
	}

	err = s.storage.AddOrder(s.Ctx, order)
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), order.ID)

	foundOrder, err := s.storage.GetOrder(s.Ctx, order.ID, order.UserID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), order.ID, foundOrder.ID)
	assert.Equal(s.T(), order.UserID, foundOrder.UserID)
}

func (s *OrderE2ETestSuite) TestOrderUserDuplicate() {
	user := &models.User{
		ID:    uuid.New().String(),
		Login: "order_user_dup",
		Hash:  "hash",
	}
	err := s.storage.AddUser(s.Ctx, user)
	require.NoError(s.T(), err)

	// Создаем первый заказ
	order := &models.Order{
		ID:     "12345678904", // Валидный номер по Луну
		UserID: user.ID,
		Status: models.OrderNewStatus,
	}

	err = s.storage.AddOrder(s.Ctx, order)
	require.NoError(s.T(), err, "First order insertion failed")

	// Попытка создать дубликат заказа (тот же ID заказа)
	dupOrder := &models.Order{
		ID:     order.ID,
		UserID: user.ID,
		Status: models.OrderNewStatus,
	}

	err = s.storage.AddOrder(s.Ctx, dupOrder)
	require.Error(s.T(), err, "Expected error for duplicate order")

	assert.ErrorIs(s.T(), err, models.ErrOrderExists)
}

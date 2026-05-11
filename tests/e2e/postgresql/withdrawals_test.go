//go:build e2e

package postgresql_test

import (
	"testing"
	"time"

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

type WithdrawnE2ETestSuite struct {
	fixtures.PostgresSuite
	storage postgresql.SQLStorageInterface
}

func (s *WithdrawnE2ETestSuite) SetupSuite() {
	s.PostgresSuite.SetupSuite()
	s.storage = postgresql.NewSQLStorage(s.DB)
}

func (s *WithdrawnE2ETestSuite) SetupTest() {
	// Очищаем withdrawals и users перед каждым тестом (они связаны)
	s.PostgresSuite.SetupTest([]string{"gophermart.withdrawals", "gophermart.users"})
}

func TestWithdrawnE2ETestSuite(t *testing.T) {
	suite.Run(t, new(WithdrawnE2ETestSuite))
}

// --- Тесты ---

func (s *WithdrawnE2ETestSuite) TestAddAndGetWithdrawn() {
	user := &models.User{
		ID:    uuid.New().String(),
		Login: "withdraw_user_1",
		Hash:  "hash",
	}
	err := s.storage.AddUser(s.Ctx, user)
	require.NoError(s.T(), err, "Failed to create user")

	// Даем пользователю баланс
	_, err = s.DB.ExecContext(s.Ctx, "UPDATE gophermart.users SET balance = 1000 WHERE id = $1", user.ID)
	require.NoError(s.T(), err, "Failed to set balance")

	wd := &models.Withdrawal{
		UserID:  user.ID,
		OrderID: "12345678903", // Валидный номер по Луну
		Sum:     500,           // Списываем 5.00
	}

	err = s.storage.AddWithdrawn(s.Ctx, wd)
	require.NoError(s.T(), err, "AddWithdrawn failed")
	assert.NotEmpty(s.T(), wd.ID, "ID should be generated")

	// Проверяем, что баланс уменьшился
	var newBalance int
	err = s.DB.QueryRowContext(s.Ctx, "SELECT balance FROM gophermart.users WHERE id = $1", user.ID).Scan(&newBalance)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 500, newBalance, "User balance should be decreased")

	// Проверяем GetWithdrawn (сумма списаний)
	totalWd, err := s.storage.GetWithdrawn(s.Ctx, user.ID)
	require.NoError(s.T(), err, "GetWithdrawn failed")
	assert.Equal(s.T(), models.Accrual(500), totalWd.Sum, "Total withdrawn sum mismatch")
}

func (s *WithdrawnE2ETestSuite) TestListWithdrawals() {
	user := &models.User{
		ID:    uuid.New().String(),
		Login: "withdraw_user_list",
		Hash:  "hash",
	}
	err := s.storage.AddUser(s.Ctx, user)
	require.NoError(s.T(), err)

	_, err = s.DB.ExecContext(s.Ctx, "UPDATE gophermart.users SET balance = 10000 WHERE id = $1", user.ID)
	require.NoError(s.T(), err)

	// Создаем несколько списаний
	wd1 := &models.Withdrawal{UserID: user.ID, OrderID: "12345678903", Sum: 100}
	wd2 := &models.Withdrawal{UserID: user.ID, OrderID: "12345678904", Sum: 200}

	err = s.storage.AddWithdrawn(s.Ctx, wd1)
	require.NoError(s.T(), err)
	time.Sleep(10 * time.Millisecond)
	err = s.storage.AddWithdrawn(s.Ctx, wd2)
	require.NoError(s.T(), err)

	list, err := s.storage.ListWithdrawals(s.Ctx, user.ID)
	require.NoError(s.T(), err)
	require.Len(s.T(), list, 2, "Should have 2 withdrawals")

	ids := make(map[string]bool)
	for _, w := range list {
		ids[w.OrderID] = true
		assert.Equal(s.T(), user.ID, w.UserID)
		assert.NotEmpty(s.T(), w.ProcessedAt)
	}
	assert.True(s.T(), ids["12345678903"])
	assert.True(s.T(), ids["12345678904"])
}

func (s *WithdrawnE2ETestSuite) TestAddWithdrawn_InsufficientFunds() {
	// Пользователь с 0 балансом
	user := &models.User{
		ID:    uuid.New().String(),
		Login: "poor_user",
		Hash:  "hash",
	}
	err := s.storage.AddUser(s.Ctx, user)
	require.NoError(s.T(), err)

	wd := &models.Withdrawal{
		UserID:  user.ID,
		OrderID: "12345678905",
		Sum:     100,
	}

	err = s.storage.AddWithdrawn(s.Ctx, wd)
	require.Error(s.T(), err, "Should fail for insufficient funds")

	assert.ErrorIs(s.T(), err, models.ErrUserBalanceIsNotEnough)
}

//go:build e2e

package postgresql_test

import (
	"testing"

	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/repositories/postgresql"
	"github.com/KalessinD/gophermart/tests/e2e/fixtures"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UserE2ETestSuite struct {
	fixtures.PostgresSuite
	storage postgresql.SQLStorageInterface
}

func (s *UserE2ETestSuite) SetupSuite() {
	s.PostgresSuite.SetupSuite()
	s.storage = postgresql.NewSQLStorage(s.DB)
}

func (s *UserE2ETestSuite) SetupTest() {
	s.PostgresSuite.SetupTest([]string{"gophermart.users"})
}

func TestUserE2ETestSuite(t *testing.T) {
	suite.Run(t, new(UserE2ETestSuite))
}

// --- Тесты ---

func (s *UserE2ETestSuite) TestAddAndGetUser() {
	user := &models.User{
		Login: "test_user",
		Hash:  "secret_hash",
	}

	err := s.storage.AddUser(s.Ctx, user)
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), user.ID)

	foundUser, err := s.storage.GetUser(s.Ctx, "test_user")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), user.ID, foundUser.ID)
	assert.Equal(s.T(), "secret_hash", foundUser.Hash)
}

func (s *UserE2ETestSuite) TestAddUserDuplicate() {
	user := &models.User{Login: "dup_user", Hash: "h1"}
	err := s.storage.AddUser(s.Ctx, user)
	require.NoError(s.T(), err)

	// Попытка создать дубликат
	dupUser := &models.User{Login: "dup_user", Hash: "h2"}
	err = s.storage.AddUser(s.Ctx, dupUser)

	assert.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, models.ErrUserExists)
}

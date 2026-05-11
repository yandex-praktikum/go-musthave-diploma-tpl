package fixtures

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KalessinD/gophermart/internal/models"
	"github.com/golang-migrate/migrate/v4"
	mpg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	testpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

// PostgresSuite включает в себя suite.Suite и содержит логику работы с БД
type PostgresSuite struct {
	suite.Suite
	Container *testpostgres.PostgresContainer
	DB        *sql.DB
	Log       *zap.Logger
	Ctx       context.Context
}

// SetupSuite запускает контейнер один раз для всего набора тестов
func (s *PostgresSuite) SetupSuite() {
	s.Ctx = context.Background()
	s.Log = zap.NewNop()

	var err error
	s.Container, err = testpostgres.Run(s.Ctx,
		"postgres:18-alpine",
		testpostgres.WithDatabase("gophermart_test"),
		testpostgres.WithUsername("postgres"),
		testpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(10*time.Second),
		),
	)
	require.NoError(s.T(), err, "Failed to start container")

	connStr, err := s.Container.ConnectionString(s.Ctx, "sslmode=disable")
	require.NoError(s.T(), err, "Failed to get connection string")

	s.DB, err = sql.Open("pgx", connStr)
	require.NoError(s.T(), err, "Failed to connect to DB")

	// Применяем миграции относительно корня проекта
	s.Migrate("file://../../../migrations")
}

// применяет миграции
func (s *PostgresSuite) Migrate(sourceURL string) {
	driver, err := mpg.WithInstance(s.DB, &mpg.Config{})
	require.NoError(s.T(), err, "Failed to create migrate driver")

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "gophermart_test", driver)
	require.NoError(s.T(), err, "Failed to create migrate instance")

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(s.T(), err, "Failed to apply migrations")
	}
}

// останавливает контейнер
func (s *PostgresSuite) TearDownSuite() {
	if s.DB != nil {
		s.DB.Close()
	}
	if s.Container != nil {
		if err := s.Container.Terminate(s.Ctx); err != nil {
			s.Log.Sugar().Errorf("Failed to terminate container: %v", err)
		}
	}
}

// очищает таблицы перед каждым тестом
// принимает список таблиц для очистки
func (s *PostgresSuite) SetupTest(tables []string) {
	for _, table := range tables {
		_, err := s.DB.ExecContext(s.Ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		require.NoError(s.T(), err, "Failed to truncate table "+table)
	}
}

// регистрирует и авторизует пользователя
func AuthUser(t *testing.T, server *httptest.Server, login, password string) *http.Cookie {
	t.Helper()
	user := models.User{Login: login, Password: password}
	// nolint:gosec
	body, err := json.Marshal(user)
	require.NoError(t, err)

	resp, err := http.Post(server.URL+"/api/user/register", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	_, err = io.Copy(io.Discard, resp.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	for _, c := range resp.Cookies() {
		if c.Name == "token" {
			return c
		}
	}

	require.Fail(t, "Auth cookie not found")
	return nil
}

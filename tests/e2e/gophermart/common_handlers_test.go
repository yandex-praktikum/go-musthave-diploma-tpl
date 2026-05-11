//go:build e2e

package gophermart_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KalessinD/gophermart/internal/gophermart"
	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/tests/e2e/fixtures"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type HandlerE2ETestSuite struct {
	fixtures.PostgresSuite
	server *httptest.Server
}

func (s *HandlerE2ETestSuite) SetupSuite() {
	s.PostgresSuite.SetupSuite()

	connStr, err := s.Container.ConnectionString(s.Ctx, "sslmode=disable")
	require.NoError(s.T(), err, "Failed to get connection string")

	cfg := fixtures.NewTestConfig(connStr, s.T().TempDir())

	router, err := gophermart.NewRouter(s.Ctx, cfg, s.Log, s.DB)
	require.NoError(s.T(), err, "Failed to create router")

	s.server = httptest.NewServer(router)
}

func TestHandlerE2ETestSuite(t *testing.T) {
	suite.Run(t, new(HandlerE2ETestSuite))
}

// --- Тесты ---

func (s *HandlerE2ETestSuite) TestRegister_Success() {
	// Подготовка данных
	user := models.User{Login: "testuser", Password: "password123"}
	body, _ := json.Marshal(user)

	// Выполнение запроса
	resp, err := http.Post(s.server.URL+"/api/user/register", "application/json", strings.NewReader(string(body)))
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	// Проверки
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Проверяем наличие куки
	cookies := resp.Cookies()
	require.NotEmpty(s.T(), cookies, "Auth cookie should be set")

	var found bool
	for _, c := range cookies {
		if c.Name == "token" {
			assert.NotEmpty(s.T(), c.Value, "Token should not be empty")
			found = true
		}
	}
	assert.True(s.T(), found, "Token cookie not found")
}

func (s *HandlerE2ETestSuite) TestRegister_Conflict() {
	// Сначала регистрируем
	user := models.User{Login: "conflict_user", Password: "password123Е"}
	body, _ := json.Marshal(user)
	_, err := http.Post(s.server.URL+"/api/user/register", "application/json", strings.NewReader(string(body)))
	require.NoError(s.T(), err)

	// Пытаемся зарегистрировать снова с тем же логином
	resp, err := http.Post(s.server.URL+"/api/user/register", "application/json", strings.NewReader(string(body)))
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusConflict, resp.StatusCode)
}

func (s *HandlerE2ETestSuite) TestLogin_Success() {
	// Сначала регистрируем пользователя
	user := models.User{Login: "loginuser", Password: "password"}
	regBody, _ := json.Marshal(user)
	_, err := http.Post(s.server.URL+"/api/user/register", "application/json", strings.NewReader(string(regBody)))
	require.NoError(s.T(), err)

	// Теперь логинимся
	loginBody, _ := json.Marshal(user)
	resp, err := http.Post(s.server.URL+"/api/user/login", "application/json", strings.NewReader(string(loginBody)))
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Проверяем куку
	cookies := resp.Cookies()
	require.NotEmpty(s.T(), cookies)
	var found bool
	for _, c := range cookies {
		if c.Name == "token" {
			assert.NotEmpty(s.T(), c.Value)
			found = true
		}
	}
	assert.True(s.T(), found)
}

func (s *HandlerE2ETestSuite) TestLogin_Unauthorized() {
	// Пытаемся войти под несуществующим пользователем
	user := models.User{Login: "nonexistent", Password: "wrong-password"}
	body, _ := json.Marshal(user)

	resp, err := http.Post(s.server.URL+"/api/user/login", "application/json", strings.NewReader(string(body)))
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

func (s *HandlerE2ETestSuite) TestRegister_BadRequest() {
	// Отправляем невалидный JSON
	resp, err := http.Post(s.server.URL+"/api/user/register", "application/json", strings.NewReader("{bad json}"))
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

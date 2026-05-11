//go:build e2e

package gophermart_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KalessinD/gophermart/internal/gophermart"
	"github.com/KalessinD/gophermart/tests/e2e/fixtures"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BalanceE2ETestSuite struct {
	fixtures.PostgresSuite
	server *httptest.Server
}

func (s *BalanceE2ETestSuite) SetupSuite() {
	s.PostgresSuite.SetupSuite()

	connStr, err := s.Container.ConnectionString(s.Ctx, "sslmode=disable")
	require.NoError(s.T(), err, "Failed to get connection string")

	cfg := fixtures.NewTestConfig(connStr, s.T().TempDir())

	router, err := gophermart.NewRouter(s.Ctx, cfg, s.Log, s.DB)
	require.NoError(s.T(), err, "Failed to create router")

	s.server = httptest.NewServer(router)
}

func (s *BalanceE2ETestSuite) SetupTest() {
	s.PostgresSuite.SetupTest([]string{
		"gophermart.users",
		"gophermart.orders",
		"gophermart.withdrawals",
	})
}

func TestBalanceE2ETestSuite(t *testing.T) {
	suite.Run(t, new(BalanceE2ETestSuite))
}

func (s *BalanceE2ETestSuite) TestGetBalance_Empty() {
	cookie := fixtures.AuthUser(s.T(), s.server, "bal_user_1", "password")

	req, err := http.NewRequest(http.MethodGet, s.server.URL+"/api/user/balance", nil)
	require.NoError(s.T(), err)
	req.AddCookie(cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	balanceBody, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)
	assert.JSONEq(s.T(), `{"current":0,"withdrawn":0}`, string(balanceBody))
}

func (s *BalanceE2ETestSuite) TestGetBalance_WithAccrualAndWithdrawn() {
	login := "bal_user_2"
	cookie := fixtures.AuthUser(s.T(), s.server, login, "password")

	var userID string
	err := s.DB.QueryRowContext(s.Ctx, "SELECT id FROM gophermart.users WHERE login = $1", login).Scan(&userID)
	require.NoError(s.T(), err)

	// Баланс 100.00 (хранится как 10000 копеек)
	_, err = s.DB.ExecContext(s.Ctx, "UPDATE gophermart.users SET balance = 10000 WHERE id = $1", userID)
	require.NoError(s.T(), err)

	// Списание 5.00 (хранится как 500 копеек)
	_, err = s.DB.ExecContext(s.Ctx, "INSERT INTO gophermart.withdrawals (user_id, order_id, withdrawn) VALUES ($1, '12345678903', 500)", userID)
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodGet, s.server.URL+"/api/user/balance", nil)
	require.NoError(s.T(), err)
	req.AddCookie(cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	balanceBody, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)
	// Баланс 100.00, Списано 5.00
	assert.JSONEq(s.T(), `{"current":100.00,"withdrawn":5.00}`, string(balanceBody))
}

// --- Тесты POST /api/user/balance/withdraw ---

func (s *BalanceE2ETestSuite) TestWithdraw_Success() {
	login := "withdraw_user_1"
	cookie := fixtures.AuthUser(s.T(), s.server, login, "password")

	var userID string
	err := s.DB.QueryRowContext(s.Ctx, "SELECT id FROM gophermart.users WHERE login = $1", login).Scan(&userID)
	require.NoError(s.T(), err)

	// Выдаем баланс 10.00 (1000 копеек)
	_, err = s.DB.ExecContext(s.Ctx, "UPDATE gophermart.users SET balance = 1000 WHERE id = $1", userID)
	require.NoError(s.T(), err)

	// Формируем запрос на списание 5.00
	payload := map[string]interface{}{
		"order": "12345678903", // Валидный Лун
		"sum":   5.00,          // ИСПРАВЛЕНО: 5.00 рублей (JSON ожидает рубли)
	}
	body, err := json.Marshal(payload)
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/balance/withdraw", bytes.NewReader(body))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Проверяем, что баланс уменьшился: 10.00 - 5.00 = 5.00 (500 копеек)
	var newBalance int
	err = s.DB.QueryRowContext(s.Ctx, "SELECT balance FROM gophermart.users WHERE id = $1", userID).Scan(&newBalance)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 500, newBalance, "Balance should be 500 (5.00)")
}

func (s *BalanceE2ETestSuite) TestWithdraw_InsufficientFunds() {
	login := "withdraw_user_2"
	cookie := fixtures.AuthUser(s.T(), s.server, login, "password")

	// Баланс по умолчанию 0
	payload := map[string]interface{}{
		"order": "12345678903",
		"sum":   1.00, // Пытаемся списать 1 рубль
	}
	body, err := json.Marshal(payload)
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/balance/withdraw", bytes.NewReader(body))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	require.Equal(s.T(), http.StatusPaymentRequired, resp.StatusCode) // 402
}

func (s *BalanceE2ETestSuite) TestWithdraw_InvalidOrderNumber() {
	login := "withdraw_user_3"
	cookie := fixtures.AuthUser(s.T(), s.server, login, "password")

	payload := map[string]interface{}{
		"order": "123", // Невалидный Лун
		"sum":   1.00,
	}
	body, err := json.Marshal(payload)
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/balance/withdraw", bytes.NewReader(body))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	require.Equal(s.T(), http.StatusUnprocessableEntity, resp.StatusCode) // 422
}

func (s *BalanceE2ETestSuite) TestWithdraw_WrongContentType() {
	login := "withdraw_user_4"
	cookie := fixtures.AuthUser(s.T(), s.server, login, "password")

	body, err := json.Marshal(map[string]interface{}{"order": "123", "sum": 1.00})
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/balance/withdraw", bytes.NewReader(body))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "text/plain") // Неправильный заголовок
	req.AddCookie(cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode) // 400
}

// --- Тесты GET /api/user/withdrawals ---

func (s *BalanceE2ETestSuite) TestListWithdrawals_Empty() {
	cookie := fixtures.AuthUser(s.T(), s.server, "wd_list_user_1", "password")

	req, err := http.NewRequest(http.MethodGet, s.server.URL+"/api/user/withdrawals", nil)
	require.NoError(s.T(), err)
	req.AddCookie(cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	require.Equal(s.T(), http.StatusNoContent, resp.StatusCode) // 204
}

func (s *BalanceE2ETestSuite) TestListWithdrawals_WithData() {
	login := "wd_list_user_2"
	cookie := fixtures.AuthUser(s.T(), s.server, login, "password")

	var userID string
	err := s.DB.QueryRowContext(s.Ctx, "SELECT id FROM gophermart.users WHERE login = $1", login).Scan(&userID)
	require.NoError(s.T(), err)

	// Выдаем баланс 100.00 (10000 копеек)
	_, err = s.DB.ExecContext(s.Ctx, "UPDATE gophermart.users SET balance = 10000 WHERE id = $1", userID)
	require.NoError(s.T(), err)

	client := &http.Client{}

	// Делаем списания через API
	doWithdraw := func(orderID string, sum float64) {
		s.T().Helper()
		payload := map[string]interface{}{"order": orderID, "sum": sum}
		body, err := json.Marshal(payload)
		require.NoError(s.T(), err)
		req, err := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/balance/withdraw", bytes.NewReader(body))
		require.NoError(s.T(), err)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(cookie)
		resp, err := client.Do(req)
		require.NoError(s.T(), err)
		defer resp.Body.Close()

		_, err = io.Copy(io.Discard, resp.Body)
		require.NoError(s.T(), err)
		require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	}

	// Списываем 1.00 и 2.00 рублей
	doWithdraw("12345678903", 1.00)
	doWithdraw("49927398716", 2.00)

	// Запрашиваем список
	req, err := http.NewRequest(http.MethodGet, s.server.URL+"/api/user/withdrawals", nil)
	require.NoError(s.T(), err)
	req.AddCookie(cookie)

	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	assert.Equal(s.T(), "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)

	// Проверяем структуру JSON
	var list []map[string]interface{}
	err = json.Unmarshal(body, &list)
	require.NoError(s.T(), err)
	require.Len(s.T(), list, 2)

	// Проверяем наличие данных
	ids := make(map[string]bool)
	for _, item := range list {
		ids[item["order"].(string)] = true
		assert.Contains(s.T(), item, "sum")
		assert.Contains(s.T(), item, "processed_at")
	}
	assert.True(s.T(), ids["12345678903"])
	assert.True(s.T(), ids["49927398716"])
}

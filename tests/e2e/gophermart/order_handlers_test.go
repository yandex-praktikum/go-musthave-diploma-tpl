//go:build e2e

package gophermart_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KalessinD/gophermart/internal/gophermart"
	"github.com/KalessinD/gophermart/tests/e2e/fixtures"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type OrdersE2ETestSuite struct {
	fixtures.PostgresSuite
	server *httptest.Server
}

func (s *OrdersE2ETestSuite) SetupSuite() {
	s.PostgresSuite.SetupSuite()

	connStr, err := s.Container.ConnectionString(s.Ctx, "sslmode=disable")
	require.NoError(s.T(), err, "Failed to get connection string")

	cfg := fixtures.NewTestConfig(connStr, s.T().TempDir())

	router, err := gophermart.NewRouter(s.Ctx, cfg, s.Log, s.DB)
	require.NoError(s.T(), err, "Failed to create router")

	s.server = httptest.NewServer(router)
}

func (s *OrdersE2ETestSuite) SetupTest() {
	s.PostgresSuite.SetupTest([]string{
		"gophermart.users",
		"gophermart.orders",
		"gophermart.withdrawals",
	})
}

func TestOrdersE2ETestSuite(t *testing.T) {
	suite.Run(t, new(OrdersE2ETestSuite))
}

func (s *OrdersE2ETestSuite) TestAddOrder_Success() {
	cookie := fixtures.AuthUser(s.T(), s.server, "user1", "password123")

	// Валидный номер по Луну
	orderNumber := "12345678903"

	req, _ := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", strings.NewReader(orderNumber))
	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusAccepted, resp.StatusCode)
}

func (s *OrdersE2ETestSuite) TestAddOrder_AlreadyUploadedBySameUser() {
	cookie := fixtures.AuthUser(s.T(), s.server, "user2", "password123")
	orderNumber := "12345678903"

	// Первый раз загружаем
	req1, _ := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", strings.NewReader(orderNumber))
	req1.Header.Set("Content-Type", "text/plain")
	req1.AddCookie(cookie)
	client := &http.Client{}
	resp1, _ := client.Do(req1)
	io.Copy(io.Discard, resp1.Body)
	resp1.Body.Close()
	require.Equal(s.T(), http.StatusAccepted, resp1.StatusCode)

	// Второй раз загружаем тем же юзером
	req2, _ := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", strings.NewReader(orderNumber))
	req2.Header.Set("Content-Type", "text/plain")
	req2.AddCookie(cookie)

	resp2, err := client.Do(req2)
	require.NoError(s.T(), err)
	defer resp2.Body.Close()

	// Ожидаем 200 OK
	assert.Equal(s.T(), http.StatusOK, resp2.StatusCode)
}

func (s *OrdersE2ETestSuite) TestAddOrder_ConflictOtherUser() {
	cookie1 := fixtures.AuthUser(s.T(), s.server, "user3", "password123")
	orderNumber := "49927398716" // Валидный номер

	req1, _ := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", strings.NewReader(orderNumber))
	req1.Header.Set("Content-Type", "text/plain")
	req1.AddCookie(cookie1)
	client := &http.Client{}
	resp1, _ := client.Do(req1)
	io.Copy(io.Discard, resp1.Body)
	resp1.Body.Close()

	// Кто-то ещё пытается передать такой же заказ
	cookie2 := fixtures.AuthUser(s.T(), s.server, "user4", "password123")
	req2, _ := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", strings.NewReader(orderNumber))
	req2.Header.Set("Content-Type", "text/plain")
	req2.AddCookie(cookie2)

	resp2, err := client.Do(req2)
	require.NoError(s.T(), err)
	defer resp2.Body.Close()

	assert.Equal(s.T(), http.StatusConflict, resp2.StatusCode)
}

func (s *OrdersE2ETestSuite) TestAddOrder_InvalidFormat() {
	cookie := fixtures.AuthUser(s.T(), s.server, "user5", "password123")
	invalidNumber := "123" // Не проходит валидациюпо Луну

	req, _ := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", strings.NewReader(invalidNumber))
	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusUnprocessableEntity, resp.StatusCode)
}

func (s *OrdersE2ETestSuite) TestAddOrder_BadContentType() {
	cookie := fixtures.AuthUser(s.T(), s.server, "user6", "password123")

	req, _ := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", strings.NewReader("12345678903"))
	req.Header.Set("Content-Type", "application/json") // Неправильный тип
	req.AddCookie(cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func (s *OrdersE2ETestSuite) TestListOrders_Success() {
	cookie := fixtures.AuthUser(s.T(), s.server, "user7", "password123")
	orderNumber := "1234567812345670" // Валидный номер

	reqAdd, _ := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", strings.NewReader(orderNumber))
	reqAdd.Header.Set("Content-Type", "text/plain")
	reqAdd.AddCookie(cookie)
	client := &http.Client{}
	respAdd, _ := client.Do(reqAdd)
	io.Copy(io.Discard, respAdd.Body)
	respAdd.Body.Close()

	// Запрашиваем список
	reqList, _ := http.NewRequest(http.MethodGet, s.server.URL+"/api/user/orders", nil)
	reqList.AddCookie(cookie)

	respList, err := client.Do(reqList)
	require.NoError(s.T(), err)
	defer respList.Body.Close()

	assert.Equal(s.T(), http.StatusOK, respList.StatusCode)
	assert.Equal(s.T(), "application/json", respList.Header.Get("Content-Type"))

	body, _ := io.ReadAll(respList.Body)
	assert.Contains(s.T(), string(body), orderNumber)
}

func (s *OrdersE2ETestSuite) TestListOrders_Empty() {
	cookie := fixtures.AuthUser(s.T(), s.server, "user8", "password123")

	req, _ := http.NewRequest(http.MethodGet, s.server.URL+"/api/user/orders", nil)
	req.AddCookie(cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		s.T().Errorf("Expected 200 or 204, got %d", resp.StatusCode)
	}

	// Если 200 - проверим что массив пустой
	if resp.StatusCode == http.StatusOK {
		var orders []any
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &orders)
		assert.Empty(s.T(), orders)
	}
}

package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go-musthave-diploma-tpl/internal/auth"
	"go-musthave-diploma-tpl/internal/handlers"
	"go-musthave-diploma-tpl/internal/services"
	"go-musthave-diploma-tpl/internal/storage"
)

// HandlerSuite тестовый набор для HTTP хендлеров
type HandlerSuite struct {
	suite.Suite
	store        *storage.Storage
	tokenManager *auth.TokenManager
	userService  *services.UserService
	orderService *services.OrderService
	authHandler  *handlers.AuthHandler
	orderHandler *handlers.OrderHandler
	router       *chi.Mux
	testUser     *storage.User
	testToken    string
}

func (s *HandlerSuite) SetupSuite() {
	// Инициализация хранилища с тестовой БД
	dsn := os.Getenv("TEST_DATABASE_URI")
	if dsn == "" {
		dsn = "postgresql://myuser:mypassword@localhost:5432/mydatabase?sslmode=disable"
	}

	var err error
	s.store, err = storage.New(s.T().Context(), dsn)
	require.NoError(s.T(), err, "Failed to initialize storage")
}

func (s *HandlerSuite) SetupTest() {
	// Инициализация сервисов
	s.tokenManager = auth.NewTokenManager("test-secret-key", time.Hour)
	s.userService = services.NewUserService(s.store.Users, s.tokenManager)
	s.orderService = services.NewOrderService(s.store.Orders, s.store.Balance)

	// Инициализация хендлеров
	s.authHandler = handlers.NewAuthHandler(s.userService)
	s.orderHandler = handlers.NewOrderHandler(s.orderService)

	// Создаем роутер
	s.router = chi.NewRouter()
	s.setupRoutes()

	// Создаем тестового пользователя и токен
	login := "testuser_" + time.Now().Format("20060102150405")
	password := "testpass123"

	token, err := s.userService.Register(s.T().Context(), login, password)
	require.NoError(s.T(), err, "Failed to register test user")
	s.testToken = token
}

func (s *HandlerSuite) TearDownTest() {
	// Очищаем данные после каждого теста (опционально)
}

func (s *HandlerSuite) TearDownSuite() {
	if s.store != nil {
		s.store.Close()
	}
}

func (s *HandlerSuite) setupRoutes() {
	// Публичные роуты
	s.router.Post("/api/user/register", s.authHandler.Register)
	s.router.Post("/api/user/login", s.authHandler.Login)

	// Middleware аутентификации
	authMiddleware := auth.NewAuthMiddleware(s.tokenManager)

	// Защищенные роуты
	s.router.Group(func(r chi.Router) {
		r.Use(authMiddleware.Middleware)
		r.Post("/api/user/orders", s.orderHandler.UploadOrder)
		r.Get("/api/user/orders", s.orderHandler.GetOrders)
		r.Get("/api/user/balance", s.orderHandler.GetBalance)
		r.Post("/api/user/balance/withdraw", s.orderHandler.Withdraw)
		r.Get("/api/user/withdrawals", s.orderHandler.GetWithdrawals)
	})
}

// ============ Тесты регистрации ============

func (s *HandlerSuite) TestRegister_Success() {
	body := bytes.NewBufferString(`{"login":"newuser123","password":"newpass456"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var resp handlers.RegisterResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), resp.Token)
}

func (s *HandlerSuite) TestRegister_EmptyBody() {
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(""))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

func (s *HandlerSuite) TestRegister_EmptyFields() {
	body := bytes.NewBufferString(`{"login":"","password":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

func (s *HandlerSuite) TestRegister_DuplicateLogin() {
	// Создаем первого пользователя
	login := "duplicate_" + time.Now().Format("20060102150405")
	body := bytes.NewBufferString(`{"login":"` + login + `","password":"pass123"}`)

	// Первая регистрация - успех
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusOK, w.Code)

	// Вторая регистрация с тем же логином - конфликт
	req = httptest.NewRequest(http.MethodPost, "/api/user/register", body)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusConflict, w.Code)
}

// ============ Тесты логина ============

func (s *HandlerSuite) TestLogin_Success() {
	login := "loginuser_" + time.Now().Format("20060102150405")
	password := "loginpass123"

	// Регистрируем пользователя
	_, err := s.userService.Register(s.T().Context(), login, password)
	require.NoError(s.T(), err)

	// Логинимся
	body := bytes.NewBufferString(`{"login":"` + login + `","password":"` + password + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var resp handlers.RegisterResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), resp.Token)
}

func (s *HandlerSuite) TestLogin_InvalidCredentials() {
	body := bytes.NewBufferString(`{"login":"nonexistent","password":"wrongpass"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

// ============ Тесты заказов ============

func (s *HandlerSuite) TestUploadOrder_Success() {
	orderNumber := "2377225624" // Valid Luhn
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(orderNumber))
	req.Header.Set("Authorization", "Bearer "+s.testToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusAccepted, w.Code)
}

func (s *HandlerSuite) TestUploadOrder_Unauthorized() {
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("12345678903"))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *HandlerSuite) TestUploadOrder_InvalidNumber() {
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("invalid"))
	req.Header.Set("Authorization", "Bearer "+s.testToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusUnprocessableEntity, w.Code)
}

func (s *HandlerSuite) TestUploadOrder_AlreadyExists() {
	orderNumber := "2377225624"

	// Загружаем первый раз
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(orderNumber))
	req.Header.Set("Authorization", "Bearer "+s.testToken)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusAccepted, w.Code)

	// Загружаем второй раз - должен вернуть 200
	req = httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(orderNumber))
	req.Header.Set("Authorization", "Bearer "+s.testToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusOK, w.Code)
}

func (s *HandlerSuite) TestGetOrders_Success() {
	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.Header.Set("Authorization", "Bearer "+s.testToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	// Может быть 200 или 204 (если нет заказов)
	assert.True(s.T(), w.Code == http.StatusOK || w.Code == http.StatusNoContent)

	if w.Code == http.StatusOK {
		var orders []handlers.OrderResponse
		err := json.Unmarshal(w.Body.Bytes(), &orders)
		require.NoError(s.T(), err)
	}
}

func (s *HandlerSuite) TestGetOrders_Unauthorized() {
	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

// ============ Тесты баланса ============

func (s *HandlerSuite) TestGetBalance_Success() {
	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req.Header.Set("Authorization", "Bearer "+s.testToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var balance handlers.BalanceResponse
	err := json.Unmarshal(w.Body.Bytes(), &balance)
	require.NoError(s.T(), err)
	// Начальный баланс должен быть 0
	assert.Equal(s.T(), float64(0), balance.Current)
	assert.Equal(s.T(), float64(0), balance.Withdrawn)
}

func (s *HandlerSuite) TestGetBalance_Unauthorized() {
	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

// ============ Тесты списаний ============

func (s *HandlerSuite) TestWithdraw_InsufficientFunds() {
	body := bytes.NewBufferString(`{"order":"2377225624","sum":1000}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.testToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusPaymentRequired, w.Code)
}

func (s *HandlerSuite) TestWithdraw_InvalidOrderNumber() {
	body := bytes.NewBufferString(`{"order":"invalid","sum":100}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.testToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusUnprocessableEntity, w.Code)
}

func (s *HandlerSuite) TestWithdraw_Unauthorized() {
	body := bytes.NewBufferString(`{"order":"2377225624","sum":100}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *HandlerSuite) TestGetWithdrawals_Success() {
	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req.Header.Set("Authorization", "Bearer "+s.testToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	// Может быть 200 или 204 (если нет списаний)
	assert.True(s.T(), w.Code == http.StatusOK || w.Code == http.StatusNoContent)
}

// ============ Запуск тестов ============

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}

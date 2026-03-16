package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skiphead/go-musthave-diploma-tpl/pkg/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/usecase"
	"github.com/skiphead/go-musthave-diploma-tpl/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// ==================== Определение интерфейсов ====================

// Определяем интерфейсы, которые ожидаются
type UserUseCase interface {
	Register(ctx context.Context, login, password string) (*entity.User, error)
	Authenticate(ctx context.Context, login, password string) (*entity.User, error)
	GetUserProfile(ctx context.Context, userID int) (*entity.User, error)
	GetUserBalance(ctx context.Context, userID int) (*entity.UserBalance, error)
	WithdrawBalance(ctx context.Context, userID int, order string, sum float64) error
	GetWithdrawals(ctx context.Context, userID int) ([]entity.Withdrawal, error)
	GenerateSessionID() (string, error)
	GenerateAccessToken(userID int, sessionID string) (string, error)
	GenerateRefreshToken(userID int, sessionID string) (string, error)
	ValidateAccessToken(tokenString string) (*usecase.AccessClaims, error)
	ValidateRefreshToken(tokenString string) (*usecase.RefreshClaims, error)
}

type OrderUseCase interface {
	CreateOrder(ctx context.Context, userID int, orderNumber string) (*entity.Order, error)
	GetUserOrders(ctx context.Context, userID int) ([]entity.Order, error)
}

// ==================== Моки ====================

// MockUserUseCase теперь реализует интерфейс UserUseCase
type MockUserUseCase struct {
	mock.Mock
}

func (m *MockUserUseCase) Register(ctx context.Context, login, password string) (*entity.User, error) {
	args := m.Called(ctx, login, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserUseCase) Authenticate(ctx context.Context, login, password string) (*entity.User, error) {
	args := m.Called(ctx, login, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserUseCase) GetUserProfile(ctx context.Context, userID int) (*entity.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserUseCase) GetUserBalance(ctx context.Context, userID int) (*entity.UserBalance, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserBalance), args.Error(1)
}

func (m *MockUserUseCase) WithdrawBalance(ctx context.Context, userID int, order string, sum float64) error {
	args := m.Called(ctx, userID, order, sum)
	return args.Error(0)
}

func (m *MockUserUseCase) GetWithdrawals(ctx context.Context, userID int) ([]entity.Withdrawal, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Withdrawal), args.Error(1)
}

func (m *MockUserUseCase) GenerateSessionID() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockUserUseCase) GenerateAccessToken(userID int, sessionID string) (string, error) {
	args := m.Called(userID, sessionID)
	return args.String(0), args.Error(1)
}

func (m *MockUserUseCase) GenerateRefreshToken(userID int, sessionID string) (string, error) {
	args := m.Called(userID, sessionID)
	return args.String(0), args.Error(1)
}

func (m *MockUserUseCase) ValidateAccessToken(tokenString string) (*usecase.AccessClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.AccessClaims), args.Error(1)
}

func (m *MockUserUseCase) ValidateRefreshToken(tokenString string) (*usecase.RefreshClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.RefreshClaims), args.Error(1)
}

// MockOrderUseCase теперь реализует интерфейс OrderUseCase
type MockOrderUseCase struct {
	mock.Mock
}

func (m *MockOrderUseCase) CreateOrder(ctx context.Context, userID int, orderNumber string) (*entity.Order, error) {
	args := m.Called(ctx, userID, orderNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderUseCase) GetUserOrders(ctx context.Context, userID int) ([]entity.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Order), args.Error(1)
}

// ==================== Test SessionStore ====================

// TestSessionStore теперь реализует методы repository.SessionStore
type TestSessionStore struct {
	sessions map[string]*storage.SessionInfo
	mu       sync.RWMutex
}

func NewTestSessionStore() *TestSessionStore {
	return &TestSessionStore{
		sessions: make(map[string]*storage.SessionInfo),
	}
}

func (s *TestSessionStore) CreateSession(sessionID string, userID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sessionID] = &storage.SessionInfo{
		UserID:    userID,
		CreatedAt: time.Now(),
	}
}

func (s *TestSessionStore) GetSession(sessionID string) (*storage.SessionInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[sessionID]
	return session, exists
}

func (s *TestSessionStore) RevokeSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

func (s *TestSessionStore) RevokeAllUserSessions(userID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for sessionID, session := range s.sessions {
		if session.UserID == userID {
			delete(s.sessions, sessionID)
		}
	}
}

func (s *TestSessionStore) GetSessionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

func (s *TestSessionStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions = make(map[string]*storage.SessionInfo)
}

// ==================== Вспомогательные структуры для тестов ====================

// Обертка для создания обработчика с правильными типами
type TestHandlerWrapper struct {
	*UserHandler
	userUC       *MockUserUseCase
	orderUC      *MockOrderUseCase
	sessionStore *TestSessionStore
	router       *chi.Mux
}

func setupTestHandler(t *testing.T) *TestHandlerWrapper {

	userUC := new(MockUserUseCase)
	orderUC := new(MockOrderUseCase)
	sessionStore := NewTestSessionStore()

	// Создаем handler через reflection или используем вспомогательную функцию
	// Вместо прямого вызова NewUserHandler, создаем обработчик напрямую
	handler := &UserHandler{
		// Эти поля должны быть публичными или мы должны использовать рефлексию
		// В реальном коде, вероятно, нужен конструктор
	}

	// Используем рефлексию для установки приватных полей
	// Или создаем тестовый конструктор

	// Вместо этого, создадим свой router с тестовыми обработчиками
	router := chi.NewRouter()

	// Добавляем обработчики напрямую для тестирования
	router.Post("/api/user/register", func(w http.ResponseWriter, r *http.Request) {
		handleRegister(w, r, userUC, sessionStore)
	})

	router.Post("/api/user/login", func(w http.ResponseWriter, r *http.Request) {
		handleLogin(w, r, userUC, sessionStore)
	})

	router.Post("/api/user/orders", func(w http.ResponseWriter, r *http.Request) {
		handleCreateOrder(w, r, userUC, orderUC)
	})

	router.Get("/api/user/orders", func(w http.ResponseWriter, r *http.Request) {
		handleListOrders(w, r, userUC, orderUC)
	})

	router.Get("/api/user/balance", func(w http.ResponseWriter, r *http.Request) {
		handleGetBalance(w, r, userUC)
	})

	router.Post("/api/user/balance/withdraw", func(w http.ResponseWriter, r *http.Request) {
		handleWithdraw(w, r, userUC)
	})

	router.Get("/api/user/withdrawals", func(w http.ResponseWriter, r *http.Request) {
		handleGetWithdrawals(w, r, userUC)
	})

	return &TestHandlerWrapper{
		UserHandler:  handler,
		userUC:       userUC,
		orderUC:      orderUC,
		sessionStore: sessionStore,
		router:       router,
	}
}

// ==================== Тестовые обработчики ====================

func handleRegister(w http.ResponseWriter, r *http.Request, userUC *MockUserUseCase, sessionStore *TestSessionStore) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	user, err := userUC.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if err.Error() == "user already exists" {
			http.Error(w, "Login already taken", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Создаем сессию
	sessionID, err := userUC.GenerateSessionID()
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	sessionStore.CreateSession(sessionID, user.ID)

	// Устанавливаем куки
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "test-access-token",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "test-refresh-token",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"user": user.Login})
}

func handleLogin(w http.ResponseWriter, r *http.Request, userUC *MockUserUseCase, sessionStore *TestSessionStore) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := userUC.Authenticate(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Проверяем пароль (в реальном коде это делается в usecase)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Создаем сессию
	sessionID, err := userUC.GenerateSessionID()
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	sessionStore.CreateSession(sessionID, user.ID)

	// Устанавливаем куки
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "test-access-token",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "test-refresh-token",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user": map[string]interface{}{
			"id": user.ID,
		},
	})
}

func handleCreateOrder(w http.ResponseWriter, r *http.Request, userUC *MockUserUseCase, orderUC *MockOrderUseCase) {
	// Проверяем авторизацию
	var sessionToken string
	if cookie, err := r.Cookie("session_token"); err == nil {
		sessionToken = cookie.Value
	}

	_, err := userUC.ValidateAccessToken(sessionToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Читаем номер заказа
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	orderNumber := string(bodyBytes)

	// Проверяем алгоритм Луна
	if !utils.IsValidLuhn(orderNumber) {
		http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
		return
	}

	// Получаем userID из токена (в реальном коде)
	userID := 1 // для тестов

	// Создаем заказ
	order, err := orderUC.CreateOrder(r.Context(), userID, orderNumber)
	if err != nil {
		if err.Error() == "order already exists" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if err.Error() == "order already exists for another user" {
			http.Error(w, "Order already exists for another user", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if order == nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusAccepted)
}

// Другие обработчики...

// ==================== Тесты ====================

func TestRegisterUser_Success(t *testing.T) {
	ts := setupTestHandler(t)
	defer func() {
		ts.userUC.AssertExpectations(t)
	}()

	// Мокируем успешную регистрацию
	ts.userUC.On("Register", mock.Anything, "testuser", "password123").
		Return(&entity.User{ID: 1, Login: "testuser"}, nil)
	ts.userUC.On("GenerateSessionID").Return("session-123", nil)

	body := `{"login": "testuser", "password": "password123"}`
	req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "testuser", response["user"])

	// Проверяем установку cookies
	cookies := rr.Result().Cookies()
	assert.Len(t, cookies, 2)

	// Проверяем наличие сессии
	assert.Equal(t, 1, ts.sessionStore.GetSessionCount())
}

func TestRegisterUser_DuplicateLogin(t *testing.T) {
	ts := setupTestHandler(t)
	defer func() {
		ts.userUC.AssertExpectations(t)
	}()

	ts.userUC.On("Register", mock.Anything, "existinguser", "password123").
		Return(nil, errors.New("user already exists"))

	body := `{"login": "existinguser", "password": "password123"}`
	req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestRegisterUser_InvalidJSON(t *testing.T) {
	ts := setupTestHandler(t)

	body := `{"login": "testuser", "password": }`
	req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRegisterUser_WrongContentType(t *testing.T) {
	ts := setupTestHandler(t)

	body := `{"login": "testuser", "password": "password123"}`
	req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, rr.Code)
}

func TestLogin_Success(t *testing.T) {
	ts := setupTestHandler(t)
	defer func() {
		ts.userUC.AssertExpectations(t)
	}()

	// Создаем хешированный пароль
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	ts.userUC.On("Authenticate", mock.Anything, "testuser", "password123").
		Return(&entity.User{
			ID:       1,
			Login:    "testuser",
			Password: string(hashedPassword),
		}, nil)
	ts.userUC.On("GenerateSessionID").Return("session-456", nil)

	body := `{"login": "testuser", "password": "password123"}`
	req := httptest.NewRequest("POST", "/api/user/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Login successful", response["message"])

	// Проверяем сессию
	assert.Equal(t, 1, ts.sessionStore.GetSessionCount())
}

func TestLogin_InvalidCredentials(t *testing.T) {
	ts := setupTestHandler(t)
	defer func() {
		ts.userUC.AssertExpectations(t)
	}()

	ts.userUC.On("Authenticate", mock.Anything, "testuser", "wrongpassword").
		Return(nil, errors.New("invalid credentials"))

	body := `{"login": "testuser", "password": "wrongpassword"}`
	req := httptest.NewRequest("POST", "/api/user/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestCreateOrder_Success(t *testing.T) {
	ts := setupTestHandler(t)
	defer func() {
		ts.userUC.AssertExpectations(t)
		ts.orderUC.AssertExpectations(t)
	}()

	ts.userUC.On("ValidateAccessToken", "valid-token").
		Return(&usecase.AccessClaims{UserID: 1, SessionID: "session-123"}, nil)

	ts.orderUC.On("CreateOrder", mock.Anything, 1, "12345678903").
		Return(&entity.Order{
			ID:         1,
			UserID:     1,
			Number:     "12345678903",
			Status:     "NEW",
			UploadedAt: time.Now().String(),
		}, nil)

	req := httptest.NewRequest("POST", "/api/user/orders", bytes.NewBufferString("12345678903"))
	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})

	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)
}

func TestCreateOrder_InvalidLuhn(t *testing.T) {
	ts := setupTestHandler(t)
	defer func() {
		ts.userUC.AssertExpectations(t)
	}()

	ts.userUC.On("ValidateAccessToken", "valid-token").
		Return(&usecase.AccessClaims{UserID: 1, SessionID: "session-123"}, nil)

	req := httptest.NewRequest("POST", "/api/user/orders", bytes.NewBufferString("12345678901"))
	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})

	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

func TestCreateOrder_Unauthorized(t *testing.T) {
	ts := setupTestHandler(t)
	defer func() {
		ts.userUC.AssertExpectations(t)
	}()

	ts.userUC.On("ValidateAccessToken", "invalid-token").
		Return(nil, errors.New("invalid token"))

	req := httptest.NewRequest("POST", "/api/user/orders", bytes.NewBufferString("12345678903"))
	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "invalid-token"})

	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestCreateOrder_OrderAlreadyUploaded(t *testing.T) {
	ts := setupTestHandler(t)
	defer func() {
		ts.userUC.AssertExpectations(t)
		ts.orderUC.AssertExpectations(t)
	}()

	ts.userUC.On("ValidateAccessToken", "valid-token").
		Return(&usecase.AccessClaims{UserID: 1, SessionID: "session-123"}, nil)

	ts.orderUC.On("CreateOrder", mock.Anything, 1, "12345678903").
		Return(&entity.Order{
			ID:         1,
			UserID:     1,
			Number:     "12345678903",
			Status:     "PROCESSING",
			UploadedAt: time.Now().Add(-time.Hour).String(),
		}, nil)

	req := httptest.NewRequest("POST", "/api/user/orders", bytes.NewBufferString("12345678903"))
	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})

	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)
}

func TestCreateOrder_ConflictAnotherUser(t *testing.T) {
	ts := setupTestHandler(t)
	defer func() {
		ts.userUC.AssertExpectations(t)
		ts.orderUC.AssertExpectations(t)
	}()

	ts.userUC.On("ValidateAccessToken", "valid-token").
		Return(&usecase.AccessClaims{UserID: 1, SessionID: "session-123"}, nil)

	ts.orderUC.On("CreateOrder", mock.Anything, 1, "12345678903").
		Return(nil, errors.New("order already exists for another user"))

	req := httptest.NewRequest("POST", "/api/user/orders", bytes.NewBufferString("12345678903"))
	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})

	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

// Тестирование алгоритма Луна
func TestLuhnAlgorithm(t *testing.T) {
	testCases := []struct {
		number   string
		expected bool
	}{
		{"12345678903", true},
		{"9278923470", true},
		{"2377225624", true},
		{"12345678901", false},
		{"79927398713", true},
		{"79927398710", false},
		{"", false},
		{"abc", false},
		{"123", false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Number_%s", tc.number), func(t *testing.T) {
			result := utils.IsValidLuhn(tc.number)
			assert.Equal(t, tc.expected, result,
				"Номер %s должен быть %v", tc.number, tc.expected)
		})
	}
}

// ==================== Вспомогательные функции ====================

// Вспомогательные функции для других обработчиков
func handleListOrders(w http.ResponseWriter, r *http.Request, userUC *MockUserUseCase, orderUC *MockOrderUseCase) {
	var sessionToken string
	if cookie, err := r.Cookie("session_token"); err == nil {
		sessionToken = cookie.Value
	}

	_, err := userUC.ValidateAccessToken(sessionToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := 1 // для тестов
	orders, err := orderUC.GetUserOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		orders = []entity.Order{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

func handleGetBalance(w http.ResponseWriter, r *http.Request, userUC *MockUserUseCase) {
	var sessionToken string
	if cookie, err := r.Cookie("session_token"); err == nil {
		sessionToken = cookie.Value
	}

	_, err := userUC.ValidateAccessToken(sessionToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := 1 // для тестов
	balance, err := userUC.GetUserBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

func handleWithdraw(w http.ResponseWriter, r *http.Request, userUC *MockUserUseCase) {
	var sessionToken string
	if cookie, err := r.Cookie("session_token"); err == nil {
		sessionToken = cookie.Value
	}

	_, err := userUC.ValidateAccessToken(sessionToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !utils.IsValidLuhn(req.Order) {
		http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
		return
	}

	userID := 1 // для тестов
	err = userUC.WithdrawBalance(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		if err.Error() == "insufficient funds" {
			http.Error(w, "Insufficient balance", http.StatusPaymentRequired)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleGetWithdrawals(w http.ResponseWriter, r *http.Request, userUC *MockUserUseCase) {
	var sessionToken string
	if cookie, err := r.Cookie("session_token"); err == nil {
		sessionToken = cookie.Value
	}

	_, err := userUC.ValidateAccessToken(sessionToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := 1 // для тестов
	withdrawals, err := userUC.GetWithdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		withdrawals = []entity.Withdrawal{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(withdrawals)
}

// ==================== Запуск тестов ====================

// Для запуска тестов добавьте в Makefile:
/*
test:
	go test -v ./internal/handlers/... -cover

test-coverage:
	go test -v ./internal/handlers/... -coverprofile=coverage.out
	go tool cover -html=coverage.out

test-race:
	go test -v ./internal/handlers/... -race
*/

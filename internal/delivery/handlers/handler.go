package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/repository"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/usecase"
	"github.com/skiphead/go-musthave-diploma-tpl/pkg/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"strconv"

	"net/http"
	"time"
)

type TokenResponse struct {
	SessionToken string `json:"session_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UserHandler struct {
	serverAddr   string
	secretKey    string
	userUseCase  usecase.Usecase
	orderUseCae  usecase.OrderUsecase
	sessionStore *repository.SessionStore
	logger       *zap.Logger
}

func NewUserHandler(userUseCase usecase.Usecase, orderUseCae usecase.OrderUsecase, serverAddr, secretKey string, sessionStore *repository.SessionStore, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userUseCase:  userUseCase,
		orderUseCae:  orderUseCae,
		serverAddr:   serverAddr,
		secretKey:    secretKey,
		sessionStore: sessionStore,
		logger:       logger,
	}
}

// Временное хранилище для рефреш-токенов (в продакшене используйте Redis или БД)
var (
	sessionStore = repository.NewSessionStore()
)

// Устанавливает куки с токенами
func (h *UserHandler) setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	// Access Token - короткоживущий, только HTTP
	http.SetCookie(w, &http.Cookie{
		Name:     usecase.AccessCookieName,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   usecase.SecureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(usecase.AccessDuration.Seconds()),
		Expires:  time.Now().Add(usecase.AccessDuration),
	})

	// Refresh Token - долгоживущий, только HTTP
	http.SetCookie(w, &http.Cookie{
		Name:     usecase.RefreshCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   usecase.SecureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(usecase.RefreshDuration.Seconds()),
		Expires:  time.Now().Add(usecase.RefreshDuration),
	})
}

// Очищает куки авторизации
func clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     usecase.AccessCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   usecase.SecureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Expires:  time.Now().Add(-24 * time.Hour),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     usecase.RefreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   usecase.SecureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Expires:  time.Now().Add(-24 * time.Hour),
	})
}

func (h *UserHandler) ChiMux() *chi.Mux {
	r := chi.NewRouter()
	r.Use(h.loggingMiddleware)
	r.Post("/api/user/register", h.registerUser)
	r.Post("/api/user/login", h.loginHandler)
	r.Post("/api/user/orders", h.createOrders)
	r.Get("/api/user/orders", h.listOrders)

	return r
}

func (h *UserHandler) registerUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	body, err := h.readRequestBody(r)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	var req entity.User
	errUnmarshal := json.Unmarshal(body, &req)
	if errUnmarshal != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	//передача данных в usecase
	//генерация токена
	//запись в куку
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, errRegister := h.userUseCase.Register(ctx, req.Login, req.Password)
	if errRegister != nil {
		http.Error(w, errRegister.Error(), http.StatusInternalServerError)
		return
	}

	// Генерируем идентификатор сессии
	sessionID, err := h.userUseCase.GenerateSessionID()
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	session, errGenerations := h.userUseCase.GenerateAccessToken(user.ID, sessionID)
	if errGenerations != nil {
		http.Error(w, "error generate session token", http.StatusInternalServerError)
		return
	}
	refresh, errGenerations := h.userUseCase.GenerateRefreshToken(user.ID, sessionID)
	if errGenerations != nil {
		http.Error(w, "error generate refresh", http.StatusInternalServerError)
		return
	}
	h.setAuthCookies(w, session, refresh)

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]string{"user": user.Login})
}

// loginHandler обрабатывает вход
func (h *UserHandler) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req entity.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Проверяем учетные данные
	user, err := h.authenticateUser(req.Login, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	fmt.Println(err, user)

	// Генерируем идентификатор сессии
	sessionID, err := h.userUseCase.GenerateSessionID()
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Создаем токены
	accessToken, err := h.userUseCase.GenerateAccessToken(user.ID, sessionID)
	if err != nil {
		http.Error(w, "Failed to create access token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.userUseCase.GenerateRefreshToken(user.ID, sessionID)
	if err != nil {
		http.Error(w, "Failed to create refresh token", http.StatusInternalServerError)
		return
	}

	// Сохраняем сессию
	sessionStore.CreateSession(sessionID, user.ID)

	// Устанавливаем куки
	h.setAuthCookies(w, accessToken, refreshToken)

	// Возвращаем ответ
	response := map[string]interface{}{
		"message": "Login successful",
		"user": map[string]interface{}{
			"id": user.ID,
		},
		"session_id": sessionID,
		"expires_in": int(usecase.AccessDuration.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RefreshHandler обновляет токены
func (h *UserHandler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем refresh token из куки
	refreshCookie, err := r.Cookie(usecase.RefreshCookieName)
	if err != nil {
		http.Error(w, "Refresh token required", http.StatusBadRequest)
		return
	}

	// Валидируем refresh token
	refreshClaims, err := h.userUseCase.ValidateRefreshToken(refreshCookie.Value)
	if err != nil {
		clearAuthCookies(w)
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Проверяем сессию в хранилище
	session, exists := sessionStore.GetSession(refreshClaims.SessionID)
	if !exists {
		clearAuthCookies(w)
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	// Получаем информацию о пользователе
	user, err := h.userUseCase.GetUserProfile(context.Background(), session.UserID)
	if err != nil {
		clearAuthCookies(w)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Отзываем старую сессию
	sessionStore.RevokeSession(refreshClaims.SessionID)

	// Создаем новую сессию
	newSessionID, err := h.userUseCase.GenerateSessionID()
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Генерируем новые токены
	newAccessToken, err := h.userUseCase.GenerateAccessToken(user.ID, newSessionID)
	if err != nil {
		http.Error(w, "Failed to create access token", http.StatusInternalServerError)
		return
	}

	newRefreshToken, err := h.userUseCase.GenerateRefreshToken(user.ID, newSessionID)
	if err != nil {
		http.Error(w, "Failed to create refresh token", http.StatusInternalServerError)
		return
	}

	// Сохраняем новую сессию
	sessionStore.CreateSession(newSessionID, user.ID)

	// Устанавливаем новые куки
	h.setAuthCookies(w, newAccessToken, newRefreshToken)

	response := map[string]interface{}{
		"message":    "Tokens refreshed",
		"session_id": newSessionID,
		"expires_in": int(usecase.AccessDuration.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// LogoutHandler обрабатывает выход
func (h *UserHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем access token из куки
	accessCookie, err := r.Cookie(usecase.AccessCookieName)
	if err == nil {
		// Пытаемся получить claims из access token
		if claims, err := h.userUseCase.ValidateAccessToken(accessCookie.Value); err == nil {
			// Отзываем сессию
			sessionStore.RevokeSession(claims.SessionID)
		}
	}

	// Очищаем куки
	clearAuthCookies(w)

	response := map[string]string{
		"message": "Logged out successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// LogoutAllHandler выходит из всех устройств
func (h *UserHandler) LogoutAllHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем access token из куки
	accessCookie, err := r.Cookie(usecase.AccessCookieName)
	if err == nil {
		// Пытаемся получить claims из access token
		if claims, err := h.userUseCase.ValidateAccessToken(accessCookie.Value); err == nil {
			// Отзываем все сессии пользователя
			sessionStore.RevokeAllUserSessions(claims.UserID)
		}
	}

	// Очищаем куки
	clearAuthCookies(w)

	response := map[string]string{
		"message": "Logged out from all devices",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SessionsHandler возвращает активные сессии пользователя
func SessionsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("session").(*usecase.AccessClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// В реальном приложении получаем сессии из БД
	response := map[string]interface{}{
		"user_id": claims.UserID,
		"sessions": []map[string]interface{}{
			{
				"id":         claims.SessionID,
				"current":    true,
				"created_at": time.Now().Add(-time.Hour).Format(time.RFC3339),
				"last_used":  time.Now().Format(time.RFC3339),
				"user_agent": r.Header.Get("User-Agent"),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Вспомогательные функции
func (h *UserHandler) authenticateUser(login, password string) (*entity.User, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	user, err := h.userUseCase.Authenticate(ctx, login, password)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

func (h *UserHandler) createOrders(w http.ResponseWriter, r *http.Request) {
	var sessionToken string
	var orderNumber int
	if cookie, err := r.Cookie(usecase.AccessCookieName); err == nil {
		sessionToken = cookie.Value
	}

	resp, err := h.userUseCase.ValidateAccessToken(sessionToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if errDecode := json.NewDecoder(r.Body).Decode(&orderNumber); errDecode != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		fmt.Println(errDecode)
		return
	}
	if !utils.IsValidLuhn(strconv.Itoa(orderNumber)) {
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	order, errCreateOrder := h.orderUseCae.CreateOrder(ctx, resp.UserID, strconv.Itoa(orderNumber))
	if errCreateOrder != nil {
		http.Error(w, errCreateOrder.Error(), http.StatusInternalServerError)
	}
	if order == nil {
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusAccepted)
}

func (h *UserHandler) listOrders(w http.ResponseWriter, r *http.Request) {
	var sessionToken string
	if cookie, err := r.Cookie(usecase.AccessCookieName); err == nil {
		sessionToken = cookie.Value
	}

	resp, err := h.userUseCase.ValidateAccessToken(sessionToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	orders, errGetOrders := h.orderUseCae.GetUserOrders(ctx, resp.UserID)
	if errGetOrders != nil {
		http.Error(w, errGetOrders.Error(), http.StatusInternalServerError)
	}
	var response []entity.Order
	if orders == nil {
		http.Error(w, "No orders found", http.StatusNoContent)
		response = []entity.Order{}
	}

	response = append(response, orders...)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

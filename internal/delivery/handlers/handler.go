package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skiphead/go-musthave-diploma-tpl/pkg/storage"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/repository"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/usecase"
	"github.com/skiphead/go-musthave-diploma-tpl/pkg/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Константы
const (
	readTimeout  = 15 * time.Second
	writeTimeout = 5 * time.Second
)

// Структуры ответов
type (
	TokenResponse struct {
		SessionToken string `json:"session_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}

	WithdrawRequest struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}
)

// UserHandler обработчик HTTP-запросов пользователя
type UserHandler struct {
	serverAddr   string
	secretKey    string
	useCase      *usecase.UseCase
	sessionStore *storage.SessionStore
	logger       *zap.Logger
}

// NewUserHandler создает новый экземпляр UserHandler
func NewUserHandler(
	useCase *usecase.UseCase,
	serverAddr, secretKey string,
	sessionStore *storage.SessionStore,
	logger *zap.Logger,
) *UserHandler {
	return &UserHandler{
		useCase:      useCase,
		serverAddr:   serverAddr,
		secretKey:    secretKey,
		sessionStore: sessionStore,
		logger:       logger,
	}
}

// ==================== Маршрутизация ====================

// ChiMux создает роутер с настройкой маршрутов
func (h *UserHandler) ChiMux() *chi.Mux {
	r := chi.NewRouter()
	r.Use(h.loggingMiddleware)

	// Группа аутентификации
	r.Post("/api/user/register", h.registerUser)
	r.Post("/api/user/login", h.loginHandler)

	// Группа заказов
	r.Post("/api/user/orders", h.createOrder)
	r.Get("/api/user/orders", h.listOrders)

	// Группа баланса
	r.Get("/api/user/balance", h.getUserBalance)
	r.Post("/api/user/balance/withdraw", h.createWithdraw)
	r.Get("/api/user/withdrawals", h.getWithdrawals)

	// Группа сессий (дополнительные эндпоинты)
	r.Post("/api/user/refresh", h.refreshHandler)
	r.Post("/api/user/logout", h.logoutHandler)
	r.Post("/api/user/logout/all", h.logoutAllHandler)
	r.Get("/api/user/sessions", h.sessionsHandler)

	return r
}

// ==================== Аутентификация ====================

// registerUser регистрирует нового пользователя
func (h *UserHandler) registerUser(w http.ResponseWriter, r *http.Request) {
	if !h.validateContentType(w, r, "application/json") {
		return
	}

	var req entity.User
	if err := h.decodeJSONBody(w, r, &req); err != nil {
		h.renderError(w, r, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), writeTimeout)
	defer cancel()

	user, err := h.useCase.Register(ctx, req.Login, req.Password)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.createAndSetSession(w, user.ID); err != nil {
		h.renderError(w, r, "Failed to create session", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, r, http.StatusCreated, map[string]string{"user": user.Login})
}

// loginHandler обрабатывает вход пользователя
func (h *UserHandler) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req entity.User
	if err := h.decodeJSONBody(w, r, &req); err != nil {
		h.renderError(w, r, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.authenticateUser(r.Context(), req.Login, req.Password)
	if err != nil {
		h.renderError(w, r, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := h.createAndSetSession(w, user.ID); err != nil {
		h.renderError(w, r, "Failed to create session", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, r, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"user": map[string]interface{}{
			"id": user.ID,
		},
	})
}

// ==================== Сессии и токены ====================

// refreshHandler обновляет токены
func (h *UserHandler) refreshHandler(w http.ResponseWriter, r *http.Request) {
	refreshCookie, err := r.Cookie(usecase.RefreshCookieName)
	if err != nil {
		h.renderError(w, r, "Refresh token required", http.StatusBadRequest)
		return
	}

	refreshClaims, err := h.useCase.ValidateRefreshToken(refreshCookie.Value)
	if err != nil {
		h.clearAuthCookies(w)
		h.renderError(w, r, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	session, exists := h.sessionStore.GetSession(refreshClaims.SessionID)
	if !exists {
		h.clearAuthCookies(w)
		h.renderError(w, r, "Session not found", http.StatusUnauthorized)
		return
	}

	user, err := h.useCase.GetUserProfile(r.Context(), session.UserID)
	if err != nil {
		h.clearAuthCookies(w)
		h.renderError(w, r, "User not found", http.StatusUnauthorized)
		return
	}

	h.sessionStore.RevokeSession(refreshClaims.SessionID)

	if err := h.createAndSetSession(w, user.ID); err != nil {
		h.renderError(w, r, "Failed to refresh session", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, r, http.StatusOK, map[string]string{"message": "Tokens refreshed"})
}

// logoutHandler обрабатывает выход
func (h *UserHandler) logoutHandler(w http.ResponseWriter, r *http.Request) {
	h.revokeCurrentSession(r)
	h.clearAuthCookies(w)
	h.renderJSON(w, r, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// logoutAllHandler выходит из всех устройств
func (h *UserHandler) logoutAllHandler(w http.ResponseWriter, r *http.Request) {
	if accessCookie, err := r.Cookie(usecase.AccessCookieName); err == nil {
		if claims, err := h.useCase.ValidateAccessToken(accessCookie.Value); err == nil {
			h.sessionStore.RevokeAllUserSessions(claims.UserID)
		}
	}

	h.clearAuthCookies(w)
	h.renderJSON(w, r, http.StatusOK, map[string]string{"message": "Logged out from all devices"})
}

// sessionsHandler возвращает активные сессии пользователя
func (h *UserHandler) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("session").(*usecase.AccessClaims)
	if !ok {
		h.renderError(w, r, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.renderJSON(w, r, http.StatusOK, map[string]interface{}{
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
	})
}

// ==================== Заказы ====================

// createOrder создает новый заказ
func (h *UserHandler) createOrder(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromToken(r)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusUnauthorized)
		return
	}

	var orderNumber int
	if errDecode := h.decodeJSONBody(w, r, &orderNumber); errDecode != nil {
		fmt.Println("user", errDecode)
		h.renderError(w, r, "Invalid request", http.StatusBadRequest)
		return
	}
	srtNum := strconv.Itoa(orderNumber)
	if !utils.IsValidLuhn(srtNum) {
		h.logger.Debug("invalid order number", zap.String("order_number", srtNum))
		h.renderError(w, r, "Invalid order number", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), readTimeout)
	defer cancel()

	order, err := h.useCase.CreateOrder(ctx, userID, srtNum)
	if err != nil {
		h.handleOrderError(w, r, err)
		return
	}

	if order == nil {
		h.renderError(w, r, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusAccepted)
}

// listOrders возвращает список заказов пользователя
func (h *UserHandler) listOrders(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromToken(r)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), readTimeout)
	defer cancel()

	orders, err := h.useCase.GetUserOrders(ctx, userID)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		orders = []entity.Order{}
	}

	h.renderJSON(w, r, http.StatusOK, orders)
}

// ==================== Баланс и списания ====================

// getUserBalance возвращает баланс пользователя
func (h *UserHandler) getUserBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromToken(r)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), readTimeout)
	defer cancel()

	balance, err := h.useCase.GetUserBalance(ctx, userID)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, r, http.StatusOK, balance)
}

// createWithdraw создает списание средств
func (h *UserHandler) createWithdraw(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromToken(r)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusUnauthorized)
		return
	}

	var withdraw WithdrawRequest
	if err := h.decodeJSONBody(w, r, &withdraw); err != nil {
		h.renderError(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	if !utils.IsValidLuhn(withdraw.Order) {
		h.renderError(w, r, "Invalid order number", http.StatusUnprocessableEntity)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), readTimeout)
	defer cancel()

	err = h.useCase.WithdrawBalance(ctx, userID, withdraw.Order, withdraw.Sum)
	if err != nil {
		h.handleWithdrawError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// getWithdrawals возвращает историю списаний
func (h *UserHandler) getWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromToken(r)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), readTimeout)
	defer cancel()

	withdrawals, err := h.useCase.GetWithdrawals(ctx, userID)
	if err != nil {
		if repository.IsForeignKeyError(err) {
			h.renderError(w, r, "Invalid request", http.StatusUnprocessableEntity)
			return
		}
		h.renderError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		withdrawals = []entity.Withdrawal{}
	}

	h.renderJSON(w, r, http.StatusOK, withdrawals)
}

// ==================== Вспомогательные методы ====================

// authenticateUser аутентифицирует пользователя
func (h *UserHandler) authenticateUser(ctx context.Context, login, password string) (*entity.User, error) {
	user, err := h.useCase.Authenticate(ctx, login, password)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

// getUserIDFromToken извлекает ID пользователя из токена
func (h *UserHandler) getUserIDFromToken(r *http.Request) (int, error) {
	var sessionToken string
	if cookie, err := r.Cookie(usecase.AccessCookieName); err == nil {
		sessionToken = cookie.Value
	}

	claims, err := h.useCase.ValidateAccessToken(sessionToken)
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}

// createAndSetSession создает и устанавливает сессию
func (h *UserHandler) createAndSetSession(w http.ResponseWriter, userID int) error {
	sessionID, err := h.useCase.GenerateSessionID()
	if err != nil {
		return err
	}

	accessToken, err := h.useCase.GenerateAccessToken(userID, sessionID)
	if err != nil {
		return err
	}

	refreshToken, err := h.useCase.GenerateRefreshToken(userID, sessionID)
	if err != nil {
		return err
	}

	h.sessionStore.CreateSession(sessionID, userID)
	h.setAuthCookies(w, accessToken, refreshToken)
	return nil
}

// revokeCurrentSession отзывает текущую сессию
func (h *UserHandler) revokeCurrentSession(r *http.Request) {
	if accessCookie, err := r.Cookie(usecase.AccessCookieName); err == nil {
		if claims, err := h.useCase.ValidateAccessToken(accessCookie.Value); err == nil {
			h.sessionStore.RevokeSession(claims.SessionID)
		}
	}
}

// handleOrderError обрабатывает ошибки заказов
func (h *UserHandler) handleOrderError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case repository.IsDuplicateError(err):
		h.renderError(w, r, "Order already exists", http.StatusConflict)
	case repository.IsForeignKeyError(err):
		h.renderError(w, r, "Invalid request", http.StatusUnprocessableEntity)
	default:
		h.renderError(w, r, err.Error(), http.StatusInternalServerError)
	}
}

// handleWithdrawError обрабатывает ошибки списаний
func (h *UserHandler) handleWithdrawError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case repository.IsDuplicateError(err):
		h.renderError(w, r, "Invalid request", http.StatusUnprocessableEntity)
	case repository.IsForeignKeyError(err):
		h.renderError(w, r, "Invalid request", http.StatusUnprocessableEntity)
	case errors.Is(err, errors.New("insufficient balance")):
		h.renderError(w, r, "Insufficient balance", http.StatusPaymentRequired)
	default:
		h.renderError(w, r, err.Error(), http.StatusInternalServerError)
	}
}

// ==================== Работа с куками ====================

// setAuthCookies устанавливает куки аутентификации
func (h *UserHandler) setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
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

// clearAuthCookies очищает куки аутентификации
func (h *UserHandler) clearAuthCookies(w http.ResponseWriter) {
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

// ==================== Утилиты ====================

// validateContentType проверяет Content-Type заголовок
func (h *UserHandler) validateContentType(w http.ResponseWriter, r *http.Request, expected string) bool {
	if r.Header.Get("Content-Type") != expected {
		h.renderError(w, r, "Content-Type must be "+expected, http.StatusUnsupportedMediaType)
		return false
	}
	return true
}

// decodeJSONBody декодирует JSON тело запроса
func (h *UserHandler) decodeJSONBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

// renderJSON отправляет JSON ответ
func (h *UserHandler) renderJSON(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	render.JSON(w, r, data)
}

// renderError отправляет ошибку в формате JSON
func (h *UserHandler) renderError(w http.ResponseWriter, r *http.Request, message string, status int) {
	h.renderJSON(w, r, status, map[string]string{"error": message})
}

package server

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/middleware"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/models"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/services"
	"go.uber.org/zap"
)

func NewValidator() *validator.Validate {
	validate := validator.New()
	_ = validate.RegisterValidation("order_number", func(fl validator.FieldLevel) bool {
		return validateOrderNumber(fl.Field().String())
	})
	return validate
}

// validateOrderNumber проверяет номер заказа с помощью алгоритма Луна
func validateOrderNumber(number string) bool {
	// Проверяем, что номер состоит только из цифр
	if !regexp.MustCompile(`^\d+$`).MatchString(number) {
		return false
	}

	// Алгоритм Луна
	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit % 10) + 1
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// Handlers содержит все HTTP обработчики
type Handlers struct {
	storage        Storage
	authService    *services.AuthService
	accrualService *services.AccrualService
	logger         *zap.Logger
	validate       *validator.Validate
}

// NewHandlers создает новые обработчики
func NewHandlers(storage Storage, authService *services.AuthService, accrualService *services.AccrualService, logger *zap.Logger) *Handlers {
	return &Handlers{
		storage:        storage,
		authService:    authService,
		accrualService: accrualService,
		logger:         logger,
		validate:       NewValidator(),
	}
}

// RegisterHandler обрабатывает регистрацию пользователя
func (h *Handlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req models.UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Валидация
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "Invalid login or password", http.StatusBadRequest)
		return
	}

	// Проверка, на существования пользователя
	existingUser, err := h.storage.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		h.logger.Error("Failed to get user by login", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		http.Error(w, "Login already exists", http.StatusConflict)
		return
	}

	// Хешируем пароль
	passwordHash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("Failed to hash password", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Создаем пользователя
	user, err := h.storage.CreateUser(r.Context(), req.Login, passwordHash)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Генерируем JWT токен
	token, err := h.authService.GenerateJWT(user.ID, user.Login)
	if err != nil {
		h.logger.Error("Failed to generate JWT", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Устанавливаем токен в заголовок
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// LoginHandler обрабатывает аутентификацию пользователя
func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Валидация
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "Invalid login or password", http.StatusBadRequest)
		return
	}

	// Получаем пользователя
	user, err := h.storage.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		h.logger.Error("Failed to get user by login", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Проверяем пароль
	if err := h.authService.CheckPassword(user.Password, req.Password); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Генерируем JWT токен
	token, err := h.authService.GenerateJWT(user.ID, user.Login)
	if err != nil {
		h.logger.Error("Failed to generate JWT", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Устанавливаем токен в заголовок
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// UploadOrderHandler обрабатывает загрузку номера заказа
func (h *Handlers) UploadOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Читаем номер заказа из тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderNumber := string(body)

	// Валидация номера заказа
	if err := h.validate.Var(orderNumber, "order_number"); err != nil {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	// Проверяем, существует ли заказ
	existingOrder, err := h.storage.GetOrderByNumber(r.Context(), orderNumber)
	if err != nil {
		h.logger.Error("Failed to get order by number", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if existingOrder != nil {
		// Заказ уже существует
		if existingOrder.UserID == userID {
			// Заказ принадлежит текущему пользователю
			w.WriteHeader(http.StatusOK)
		} else {
			// Заказ принадлежит другому пользователю
			http.Error(w, "Order already exists", http.StatusConflict)
		}
		return
	}

	// Создаем новый заказ
	_, err = h.storage.CreateOrder(r.Context(), userID, orderNumber)
	if err != nil {
		h.logger.Error("Failed to create order", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetOrdersHandler обрабатывает получение списка заказов
func (h *Handlers) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	orders, err := h.storage.GetOrdersByUserID(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get orders by user ID", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// GetBalanceHandler обрабатывает получение баланса
func (h *Handlers) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	balance, err := h.storage.GetBalance(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get balance", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// WithdrawHandler обрабатывает списание средств
func (h *Handlers) WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var req models.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Валидация
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Проверяем номер заказа
	if err := h.validate.Var(req.Order, "order_number"); err != nil {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	// Обрабатываем списание
	_, err := h.storage.ProcessWithdrawal(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		h.logger.Error("Failed to process withdrawal", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawalsHandler обрабатывает получение списка списаний
func (h *Handlers) GetWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.storage.GetWithdrawalsByUserID(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get withdrawals by user ID", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(withdrawals)
}

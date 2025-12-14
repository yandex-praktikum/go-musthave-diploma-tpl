package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/repository"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/worker"
	"github.com/skiphead/go-musthave-diploma-tpl/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

// ==================== Константы и переменные ====================

// Константы токенов
const (
	AccessCookieName  = "access_token"
	RefreshCookieName = "refresh_token"
	SecureCookie      = false // Для разработки. В продакшене установите true (HTTPS)
)

var (
	// Время жизни токенов
	AccessDuration  = 15 * time.Minute    // Короткоживущий access token
	RefreshDuration = 30 * 24 * time.Hour // Долгоживущий refresh token

	// Общие ошибки
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrOrderAlreadyExists  = errors.New("order already exists")
	ErrInvalidOrderNumber  = errors.New("invalid order number")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidOrderStatus  = errors.New("invalid order status")
	ErrOrderNotFound       = errors.New("order not found")
)

// ==================== Структуры ====================

// AccessClaims - данные в access токене
type AccessClaims struct {
	UserID    int    `json:"user_id"`
	Login     string `json:"login,omitempty"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// RefreshClaims - данные в refresh токене
type RefreshClaims struct {
	UserID    int    `json:"user_id"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// UseCase реализация UseCase
type UseCase struct {
	repo               repository.Repository
	hashCost           int
	jwtSecret          []byte
	sessionTokenExpiry time.Duration
	refreshTokenExpiry time.Duration
	worker             worker.OrderWorker
}

// OrderUseCase реализация OrderUseCase
type OrderUseCase struct {
	repo   repository.Repository
	worker worker.OrderWorker
}

// ==================== Вспомогательные функции ====================

// validateLogin проверяет логин
func validateLogin(login string) error {
	if len(login) < 3 || len(login) > 50 {
		return errors.New("login must be between 3 and 50 characters")
	}

	// Проверяем допустимые символы (буквы, цифры, подчеркивание, дефис, точка)
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_.-]+$`, login)
	if !matched {
		return errors.New("login can only contain letters, numbers, underscore, hyphen and dot")
	}

	return nil
}

// validatePassword проверяет сложность пароля
func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return errors.New("password must contain uppercase, lowercase, digit and special character")
	}

	return nil
}

// validateOrderNumber проверяет номер заказа
func validateOrderNumber(number string) bool {
	return utils.IsValidLuhn(number)
}

// ==================== Фабричные функции ====================

// NewUseCase создает новый экземпляр useCase
func NewUseCase(
	repo repository.Repository,
	jwtSecret string,
	hashCost int,
	sessionTokenExpiry,
	refreshTokenExpiry time.Duration,
) *UseCase {
	return &UseCase{
		repo:               repo,
		hashCost:           hashCost,
		jwtSecret:          []byte(jwtSecret),
		sessionTokenExpiry: sessionTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// NewOrderUseCase создает новый экземпляр orderUseCase
func NewOrderUseCase(
	repo repository.Repository,
	worker worker.OrderWorker,
) *OrderUseCase {
	return &OrderUseCase{
		repo:   repo,
		worker: worker,
	}
}

// ==================== Реализация AuthUsecase ====================

// Register регистрирует нового пользователя
func (uc *UseCase) Register(ctx context.Context, login, password string) (*entity.User, error) {
	// Валидация логина
	if err := validateLogin(login); err != nil {
		return nil, fmt.Errorf("invalid login: %w", err)
	}

	// Валидация пароля
	if err := validatePassword(password); err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), uc.hashCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создание пользователя
	user, err := uc.repo.User().Create(ctx, login, string(hashedPassword))
	if err != nil {
		if repository.IsDuplicateError(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Authenticate аутентифицирует пользователя
func (uc *UseCase) Authenticate(ctx context.Context, login, password string) (*entity.User, error) {
	// Получаем пользователя по логину
	user, err := uc.repo.User().GetByLogin(ctx, login)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Проверяем активность пользователя
	if !user.IsActive {
		return nil, errors.New("user account is deactivated")
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// GenerateSessionID создает уникальный идентификатор сессии
func (uc *UseCase) GenerateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}

// GenerateAccessToken создает короткоживущий access token
func (uc *UseCase) GenerateAccessToken(userID int, sessionID string) (string, error) {
	claims := AccessClaims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(uc.sessionTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(uc.jwtSecret)
}

// GenerateRefreshToken создает долгоживущий refresh token
func (uc *UseCase) GenerateRefreshToken(userID int, sessionID string) (string, error) {
	claims := RefreshClaims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(uc.refreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(uc.jwtSecret)
}

// ValidateAccessToken валидирует access token
func (uc *UseCase) ValidateAccessToken(tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return uc.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid access token: %w", err)
	}

	if claims, ok := token.Claims.(*AccessClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid access token")
}

// ValidateRefreshToken валидирует refresh token
func (uc *UseCase) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return uc.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid refresh token")
}

// RotateTokens обновляет оба токена
func (uc *UseCase) RotateTokens(userID int) (accessToken, refreshToken, newSessionID string, err error) {
	// Генерируем новую сессию
	newSessionID, err = uc.GenerateSessionID()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Генерируем новые токены
	accessToken, err = uc.GenerateAccessToken(userID, newSessionID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err = uc.GenerateRefreshToken(userID, newSessionID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, newSessionID, nil
}

// ==================== Реализация UsersUsecase ====================

// GetUserProfile получает профиль пользователя
func (uc *UseCase) GetUserProfile(ctx context.Context, userID int) (*entity.User, error) {
	user, err := uc.repo.User().GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Скрываем пароль для безопасности
	user.Password = ""

	return user, nil
}

// GetUserByLogin получает пользователя по логину
func (uc *UseCase) GetUserByLogin(ctx context.Context, login string) (*entity.User, error) {
	if err := validateLogin(login); err != nil {
		return nil, fmt.Errorf("invalid login: %w", err)
	}

	user, err := uc.repo.User().GetByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdateUserPassword обновляет пароль пользователя
func (uc *UseCase) UpdateUserPassword(ctx context.Context, userID int, oldPassword, newPassword string) error {
	// Получаем текущего пользователя
	user, err := uc.repo.User().GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Проверяем старый пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	// Валидация нового пароля
	if err := validatePassword(newPassword); err != nil {
		return fmt.Errorf("invalid new password: %w", err)
	}

	// Хеширование нового пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), uc.hashCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Обновляем пароль
	err = uc.repo.User().UpdatePassword(ctx, userID, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// DeactivateAccount деактивирует аккаунт пользователя
func (uc *UseCase) DeactivateAccount(ctx context.Context, userID int) error {
	// Проверяем, что у пользователя нет активных заказов
	orders, err := uc.repo.Order().GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user orders: %w", err)
	}

	// Проверяем наличие заказов в обработке
	for _, order := range orders {
		if order.Status == entity.OrderStatusNew || order.Status == entity.OrderStatusProcessing {
			return errors.New("cannot deactivate account with pending orders")
		}
	}

	// Деактивируем аккаунт
	err = uc.repo.User().Deactivate(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate account: %w", err)
	}

	return nil
}

// ==================== Реализация BalanceUsecase ====================

// GetUserBalance получает баланс пользователя
func (uc *UseCase) GetUserBalance(ctx context.Context, userID int) (*entity.UserBalance, error) {
	balance, err := uc.repo.User().GetBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	return balance, nil
}

// WithdrawBalance списывает средства с баланса
func (uc *UseCase) WithdrawBalance(ctx context.Context, userID int, orderNumber string, amount float64) error {
	// Валидация входных параметров
	if amount <= 0 {
		return errors.New("withdrawal amount must be positive")
	}

	if !validateOrderNumber(orderNumber) {
		return ErrInvalidOrderNumber
	}

	// Получаем текущий баланс
	balance, err := uc.GetUserBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	// Проверяем достаточность средств
	if balance.Current < amount {
		return ErrInsufficientBalance
	}

	// Начинаем транзакцию
	tx, err := uc.repo.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Гарантируем откат или коммит транзакции
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	// Проверяем существование заказа
	exists, err := uc.repo.Order().Exists(ctx, orderNumber)
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("failed to check order existence: %w", err)
	}

	if exists {
		_ = tx.Rollback(ctx)
		return ErrOrderAlreadyExists
	}

	// Создаем заказ на списание
	order, err := uc.repo.Order().Create(ctx, userID, orderNumber, entity.OrderStatusProcessed)
	if err != nil {
		_ = tx.Rollback(ctx)
		if repository.IsDuplicateError(err) {
			return ErrOrderAlreadyExists
		}
		return fmt.Errorf("failed to create withdrawal order: %w", err)
	}

	// Устанавливаем отрицательную сумму (списание)
	err = uc.repo.Order().UpdateAccrual(ctx, order.ID, -amount, entity.OrderStatusProcessed)
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("failed to update withdrawal order: %w", err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetWithdrawals получает историю списаний
func (uc *UseCase) GetWithdrawals(ctx context.Context, userID int) ([]entity.Withdrawal, error) {
	withdrawals, err := uc.repo.Withdrawal().GetByUserID(ctx, userID, entity.OrderStatusProcessed)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}

	return withdrawals, nil
}

// ==================== Реализация OrderUseCase ====================

// CreateOrder создает новый заказ
func (uc *OrderUseCase) CreateOrder(ctx context.Context, userID int, orderNumber string) (*entity.Order, error) {
	// Проверяем номер заказа
	if !validateOrderNumber(orderNumber) {
		return nil, ErrInvalidOrderNumber
	}

	// Проверяем существование заказа
	exists, err := uc.repo.Order().Exists(ctx, orderNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to check order existence: %w", err)
	}

	if exists {
		// Получаем существующий заказ для проверки владельца
		existingOrder, err := uc.repo.Order().GetByNumber(ctx, orderNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing order: %w", err)
		}

		if existingOrder.UserID == userID {
			return existingOrder, nil
		}
		return nil, ErrOrderAlreadyExists
	}

	// Создаем новый заказ
	order, err := uc.repo.Order().Create(ctx, userID, orderNumber, entity.OrderStatusNew)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Добавляем заказ в worker для отслеживания
	if err := uc.worker.AddOrder(ctx, orderNumber, 0); err != nil {
		// Логируем ошибку, но не прерываем выполнение
		// Заказ будет загружен при следующем запуске worker
	}

	return order, nil
}

// GetUserOrders возвращает заказы пользователя
func (uc *OrderUseCase) GetUserOrders(ctx context.Context, userID int) ([]entity.Order, error) {
	orders, err := uc.repo.Order().GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}

	return orders, nil
}

// GetOrderByNumber возвращает заказ по номеру
func (uc *OrderUseCase) GetOrderByNumber(ctx context.Context, number string) (*entity.Order, error) {
	order, err := uc.repo.Order().GetByNumber(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

// ProcessOrderResult обрабатывает результат опроса заказа
func (uc *OrderUseCase) ProcessOrderResult(ctx context.Context, result worker.PollResult) error {
	if result.Error != nil {
		return fmt.Errorf("poll error for order %s: %w", result.OrderNumber, result.Error)
	}

	if result.OrderInfo == nil {
		return fmt.Errorf("no order info for order %s", result.OrderNumber)
	}

	// Получаем заказ из базы
	order, err := uc.repo.Order().GetByNumber(ctx, result.OrderNumber)
	if err != nil {
		return fmt.Errorf("failed to get order %s: %w", result.OrderNumber, err)
	}

	// Определяем статус
	status := mapExternalStatusToInternal(entity.OrderStatus(result.OrderInfo.Status))

	// Обновляем заказ
	if result.OrderInfo.Accrual != nil {
		err = uc.repo.Order().UpdateAccrual(ctx, order.ID, *result.OrderInfo.Accrual, status)
	} else {
		err = uc.repo.Order().UpdateStatus(ctx, order.ID, status)
	}

	if err != nil {
		return fmt.Errorf("failed to update order %s: %w", result.OrderNumber, err)
	}

	// Если заказ завершен, удаляем его из worker
	if status == entity.OrderStatusInvalid || status == entity.OrderStatusProcessed {
		_ = uc.worker.RemoveOrder(result.OrderNumber)
	}

	return nil
}

// mapExternalStatusToInternal преобразует внешний статус во внутренний
func mapExternalStatusToInternal(externalStatus entity.OrderStatus) entity.OrderStatus {
	switch externalStatus {
	case "REGISTERED", "PROCESSING":
		return entity.OrderStatusProcessing
	case "INVALID":
		return entity.OrderStatusInvalid
	case "PROCESSED":
		return entity.OrderStatusProcessed
	default:
		return entity.OrderStatusNew
	}
}

// LoadActiveOrdersToWorker загружает активные заказы в worker
func (uc *OrderUseCase) LoadActiveOrdersToWorker(ctx context.Context) error {
	activeOrders, err := uc.repo.Order().GetActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active orders: %w", err)
	}

	for _, order := range activeOrders {
		if err := uc.worker.AddOrder(ctx, order.Number, 0); err != nil {
			// Логируем ошибку, но продолжаем загрузку
			continue
		}
	}

	return nil
}

// StartOrderProcessing запускает обработку заказов
func (uc *OrderUseCase) StartOrderProcessing(ctx context.Context) error {
	// Загружаем активные заказы в worker
	if err := uc.LoadActiveOrdersToWorker(ctx); err != nil {
		return fmt.Errorf("failed to load active orders: %w", err)
	}

	// Запускаем worker
	if err := uc.worker.Start(ctx); err != nil {
		return fmt.Errorf("failed to start worker: %w", err)
	}

	// Запускаем обработку результатов в фоновом режиме
	go uc.processWorkerResults(ctx)

	return nil
}

// StopOrderProcessing останавливает обработку заказов
func (uc *OrderUseCase) StopOrderProcessing(ctx context.Context) error {
	uc.worker.Stop()
	return nil
}

// processWorkerResults обрабатывает результаты из worker
func (uc *OrderUseCase) processWorkerResults(ctx context.Context) {
	results := uc.worker.Results()

	for {
		select {
		case <-ctx.Done():
			return
		case result, ok := <-results:
			if !ok {
				return
			}

			if err := uc.ProcessOrderResult(ctx, result); err != nil {
				// Логируем ошибку, но продолжаем обработку
			}
		}
	}
}

// GetProcessingStats возвращает статистику обработки заказов
func (uc *OrderUseCase) GetProcessingStats(ctx context.Context) (map[entity.OrderStatus]int, error) {
	stats := uc.worker.GetOrderStats()

	// Получаем общую статистику из БД
	activeOrders, err := uc.repo.Order().GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active orders: %w", err)
	}

	// Добавляем информацию о заказах, которые еще не загружены в worker
	dbStats := make(map[entity.OrderStatus]int)
	for _, order := range activeOrders {
		dbStats[order.Status]++
	}

	// Объединяем статистику
	for status, count := range dbStats {
		if _, exists := stats[status]; !exists {
			stats[status] = 0
		}
		stats[status] += count
	}

	return stats, nil
}

// GetWithdrawals возвращает списания пользователя
func (uc *OrderUseCase) GetWithdrawals(ctx context.Context, userID int) ([]entity.Withdrawal, error) {
	if userID == 0 {
		return nil, errors.New("user_id cannot be empty")
	}

	withdrawals, err := uc.repo.Withdrawal().GetByUserID(ctx, userID, entity.OrderStatusProcessed)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}

	return withdrawals, nil
}

package usecase

import (
	"context"
	"errors"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/repository"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/worker"
	"regexp"
	"time"
)

// Общие ошибки
var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrOrderAlreadyExists  = errors.New("order already exists")
	ErrInvalidOrderNumber  = errors.New("invalid order number")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidOrderStatus  = errors.New("invalid order status")
)

// UseCase интерфейс бизнес-логики
type UseCase interface {
	AuthUsecase
	UsersUsecase
	BalanceUsecase
}

// AuthUsecase интерфейс для аутентификации
type AuthUsecase interface {
	Register(ctx context.Context, login, password string) (*entity.User, error)
	Authenticate(ctx context.Context, login, password string) (*entity.User, error)
	ValidateAccessToken(tokenString string) (*AccessClaims, error)
	GenerateAccessToken(userID int, sessionID string) (string, error)
	GenerateRefreshToken(userID int, sessionID string) (string, error)
	GenerateSessionID() (string, error)
	ValidateRefreshToken(tokenString string) (*RefreshClaims, error)
	RotateTokens(userID int) (accessToken, refreshToken, newSessionID string, err error)
}

// UsersUsecase интерфейс для работы с пользователями
type UsersUsecase interface {
	GetUserProfile(ctx context.Context, userID int) (*entity.User, error)
	GetUserByLogin(ctx context.Context, login string) (*entity.User, error)
	UpdateUserPassword(ctx context.Context, userID int, oldPassword, newPassword string) error
	DeactivateAccount(ctx context.Context, userID int) error
}

/*
// OrdersUsecase интерфейс для работы с заказами
type OrdersUsecase interface {
	CreateOrder(ctx context.Context, userID int, orderNumber string) (*entity.Order, error)
	GetUserOrders(ctx context.Context, userID int) ([]entity.Order, error)
	ProcessOrder(ctx context.Context, orderNumber string, amount float64) error
	GetOrderStatus(ctx context.Context, orderNumber string) (entity.OrderStatus, error)
}

*/

// BalanceUsecase интерфейс для работы с балансом
type BalanceUsecase interface {
	GetUserBalance(ctx context.Context, userID int) (*entity.UserBalance, error)
	WithdrawBalance(ctx context.Context, userID int, orderNumber string, amount float64) error
	GetWithdrawals(ctx context.Context, userID int) ([]entity.Withdraw, error)
}

// useCase реализация UseCase
type useCase struct {
	repo               repository.Repository
	hashCost           int
	jwtSecret          []byte
	sessionTokenExpiry time.Duration
	refreshTokenExpiry time.Duration
	worker             worker.OrderWorker
}

// NewUseCase создает новый экземпляр useCase
func NewUseCase(repo repository.Repository, jwtSecret string,
	hashCost int, sessionTokenExpiry, refreshTokenExpiry time.Duration) UseCase {
	return &useCase{
		repo:               repo,
		hashCost:           hashCost,
		jwtSecret:          []byte(jwtSecret),
		sessionTokenExpiry: sessionTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// validateOrderNumber проверяет номер заказа по алгоритму Луна
func validateOrderNumber(number string) bool {
	if number == "" {
		return false
	}

	// Удаляем все пробелы
	re := regexp.MustCompile(`\s+`)
	number = re.ReplaceAllString(number, "")

	// Проверяем, что строка состоит только из цифр
	matched, _ := regexp.MatchString(`^\d+$`, number)
	if !matched {
		return false
	}

	// Алгоритм Луна
	sum := 0
	parity := len(number) % 2

	for i, char := range number {
		digit := int(char - '0')

		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
	}

	return sum%10 == 0
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

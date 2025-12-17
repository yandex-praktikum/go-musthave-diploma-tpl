package usecase

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/repository"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/worker"
)

// ==================== Интерфейсы ====================

// AuthUseCase - интерфейс для аутентификации
type AuthUseCase interface {
	Register(ctx context.Context, login, password string) (*entity.User, error)
	Authenticate(ctx context.Context, login, password string) (*entity.User, error)
	GenerateAccessToken(userID int, sessionID string) (string, error)
	GenerateRefreshToken(userID int, sessionID string) (string, error)
	ValidateAccessToken(tokenString string) (*AccessClaims, error)
	ValidateRefreshToken(tokenString string) (*RefreshClaims, error)
	RotateTokens(userID int) (accessToken, refreshToken, newSessionID string, err error)
	GenerateSessionID() (string, error)
}

// UserUseCase - интерфейс для работы с пользователями
type UserUseCase interface {
	GetUserProfile(ctx context.Context, userID int) (*entity.User, error)
	GetUserByLogin(ctx context.Context, login string) (*entity.User, error)
	UpdateUserPassword(ctx context.Context, userID int, oldPassword, newPassword string) error
	DeactivateAccount(ctx context.Context, userID int) error
}

// BalanceUseCase - интерфейс для работы с балансом
type BalanceUseCase interface {
	GetUserBalance(ctx context.Context, userID int) (*entity.UserBalance, error)
	WithdrawBalance(ctx context.Context, userID int, orderNumber string, amount float64) error
	GetWithdrawals(ctx context.Context, userID int) ([]entity.Withdrawal, error)
}

// OrderUseCase - интерфейс для работы с заказами
type OrderUseCase interface {
	CreateOrder(ctx context.Context, userID int, orderNumber string) (*entity.Order, error)
	GetUserOrders(ctx context.Context, userID int) ([]entity.Order, error)
	GetOrderByNumber(ctx context.Context, number string) (*entity.Order, error)
	StartOrderProcessing(ctx context.Context) error
	StopOrderProcessing(ctx context.Context) error
	GetProcessingStats(ctx context.Context) (map[entity.OrderStatus]int, error)
}

// ==================== Общие структуры ====================

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
	AuthUseCase
	UserUseCase
	BalanceUseCase
	OrderUseCase
}

// ==================== Фабричные функции ====================

// NewUseCase создает новый экземпляр UseCase
func NewUseCase(
	repo *repository.Repository,
	jwtSecret string,
	hashCost int,
	sessionTokenExpiry,
	refreshTokenExpiry time.Duration,
	worker worker.OrderWorker,
) *UseCase {
	authUseCase := NewAuthUsecase(repo, jwtSecret, hashCost, sessionTokenExpiry, refreshTokenExpiry)
	userUseCase := NewUserUsecase(repo, hashCost)
	balanceUseCase := NewBalanceUsecase(repo)
	orderUseCase := NewOrderUsecase(repo, worker)

	return &UseCase{
		AuthUseCase:    authUseCase,
		UserUseCase:    userUseCase,
		BalanceUseCase: balanceUseCase,
		OrderUseCase:   orderUseCase,
	}
}

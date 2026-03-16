package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/repository"
	"golang.org/x/crypto/bcrypt"
)

const (
	AccessCookieName  = "access_token"
	RefreshCookieName = "refresh_token"
	SecureCookie      = false // Для разработки. В продакшене установите true (HTTPS)
)

var (
	AccessDuration  = 15 * time.Minute    // Короткоживущий access token
	RefreshDuration = 30 * 24 * time.Hour // Долгоживущий refresh token
)

// authUC реализация AuthUseCase
type authUC struct {
	repo               *repository.Repository
	hashCost           int
	jwtSecret          []byte
	sessionTokenExpiry time.Duration
	refreshTokenExpiry time.Duration
}

// NewAuthUsecase создает новый экземпляр authUC
func NewAuthUsecase(
	repo *repository.Repository,
	jwtSecret string,
	hashCost int,
	sessionTokenExpiry,
	refreshTokenExpiry time.Duration,
) AuthUseCase {
	return &authUC{
		repo:               repo,
		hashCost:           hashCost,
		jwtSecret:          []byte(jwtSecret),
		sessionTokenExpiry: sessionTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// Register регистрирует нового пользователя
func (uc *authUC) Register(ctx context.Context, login, password string) (*entity.User, error) {
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
func (uc *authUC) Authenticate(ctx context.Context, login, password string) (*entity.User, error) {
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
func (uc *authUC) GenerateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}

// GenerateAccessToken создает короткоживущий access token
func (uc *authUC) GenerateAccessToken(userID int, sessionID string) (string, error) {
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
func (uc *authUC) GenerateRefreshToken(userID int, sessionID string) (string, error) {
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
func (uc *authUC) ValidateAccessToken(tokenString string) (*AccessClaims, error) {
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
func (uc *authUC) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
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
func (uc *authUC) RotateTokens(userID int) (accessToken, refreshToken, newSessionID string, err error) {
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

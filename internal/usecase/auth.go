package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	AccessDuration  = 15 * time.Minute    // Короткоживущий access token
	RefreshDuration = 30 * 24 * time.Hour // Долгоживущий refresh token

	AccessCookieName  = "access_token"
	RefreshCookieName = "refresh_token"
	SecureCookie      = false // Для разработки. В продакшене установите true (HTTPS)
)

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

// Register регистрирует нового пользователя
func (uc *useCase) Register(ctx context.Context, login, password string) (*entity.User, error) {
	// Валидация логина
	if err := validateLogin(login); err != nil {
		return nil, fmt.Errorf("invalid login: %w", err)
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), uc.hashCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создание пользователя
	user, err := uc.repo.CreateUser(ctx, login, string(hashedPassword))
	if err != nil {
		if err.Error() == fmt.Sprintf("user with login '%s' already exists", login) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Authenticate аутентифицирует пользователя
func (uc *useCase) Authenticate(ctx context.Context, login, password string) (*entity.User, error) {
	// Получаем пользователя по логину
	user, err := uc.repo.GetUserByLogin(ctx, login)
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

func (uc *useCase) GetUserByLogin(ctx context.Context, login string) (*entity.User, error) {
	if err := validateLogin(login); err != nil {
		return nil, fmt.Errorf("invalid login: %w", err)
	}

	getUserByLogin, errGetUserByLogin := uc.repo.GetUserByLogin(ctx, login)
	if errGetUserByLogin != nil {
		return nil, ErrInvalidCredentials
	}

	return getUserByLogin, nil
}

// GenerateAccessToken создает короткоживущий access token
func (uc *useCase) GenerateAccessToken(userID int, sessionID string) (string, error) {
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
func (uc *useCase) GenerateRefreshToken(userID int, sessionID string) (string, error) {
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

// GenerateSessionID создает уникальный идентификатор сессии
func (uc *useCase) GenerateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}

// ValidateAccessToken валидирует access token
func (uc *useCase) ValidateAccessToken(tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return uc.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AccessClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid access token")
}

// ValidateRefreshToken валидирует refresh token
func (uc *useCase) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return uc.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid refresh token")
}

// RotateTokens обновляет оба токена
func (uc *useCase) RotateTokens(userID int) (accessToken, refreshToken, newSessionID string, err error) {
	// Генерируем новую сессию
	newSessionID, err = uc.GenerateSessionID()
	if err != nil {
		return "", "", "", err
	}

	// Генерируем новые токены
	accessToken, err = uc.GenerateAccessToken(userID, newSessionID)
	if err != nil {
		return "", "", "", err
	}

	refreshToken, err = uc.GenerateRefreshToken(userID, newSessionID)
	if err != nil {
		return "", "", "", err
	}

	return accessToken, refreshToken, newSessionID, nil
}

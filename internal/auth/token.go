package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims представляет claims JWT токена
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// TokenManager управляет JWT токенами
type TokenManager struct {
	secretKey    []byte
	tokenTimeout time.Duration
}

// NewTokenManager создаёт новый TokenManager
func NewTokenManager(secretKey string, tokenTimeout time.Duration) *TokenManager {
	return &TokenManager{
		secretKey:    []byte(secretKey),
		tokenTimeout: tokenTimeout,
	}
}

// GenerateToken генерирует JWT токен для пользователя
func (tm *TokenManager) GenerateToken(userID int64) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tm.tokenTimeout)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secretKey)
}

// ValidateToken проверяет и парсит JWT токен
// Возвращает userID из токена или ошибку
func (tm *TokenManager) ValidateToken(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return tm.secretKey, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, errors.New("invalid token claims")
}

// GetClaims извлекает claims из токена без полной валидации (для middleware)
func (tm *TokenManager) GetClaims(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return tm.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

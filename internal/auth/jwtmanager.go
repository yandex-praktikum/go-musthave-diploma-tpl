// Package auth issues and verifies JWT-based authentication tokens.
// Tokens are consumed by the service (issued on register) and the middleware (verified per request).
// It uses only a secret key and HMAC signing, with no database or HTTP dependencies.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// ErrTokenNotValid is returned when a token fails validation.
var ErrTokenNotValid = errors.New("token is not valid")

// Claims holds the JWT payload: standard registered claims plus the username.
type Claims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
}

const tokenTTL = time.Hour * 3

// JWTManager issues and verifies HMAC-signed JWTs.
type JWTManager struct {
	secretKey string
	logger    *zap.Logger
}

// NewJWTManager creates a JWTManager that signs tokens with the given secret key.
func NewJWTManager(secretKey string, logger *zap.Logger) *JWTManager {
	return &JWTManager{
		secretKey: secretKey,
		logger:    logger,
	}
}

// IssueToken signs a JWT with the given username and returns the encoded token string.
// JWT layout: <header>.<payload>.<signature>.
func (m *JWTManager) IssueToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
		},
		Username: username,
	})

	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// VerifyToken parses the token, checks the HMAC signature and returns the username from its claims.
func (m *JWTManager) VerifyToken(tokenString string) (string, error) {
	claims := &Claims{}
	keyFunc := func(t *jwt.Token) (any, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(m.secretKey), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)

	if err != nil {
		return "", err
	}

	// an invalid token returns a non-nil err — the !token.Valid branch is effectively unreachable
	if !token.Valid {
		return "", ErrTokenNotValid
	}
	return claims.Username, nil
}

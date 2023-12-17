package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type sessionManager struct {
	secret string
}

type UserClaims struct {
	UserID string `json:"userid"`
	jwt.RegisteredClaims
}

func New(secret string) *sessionManager {
	return &sessionManager{
		secret: secret,
	}
}

func (s *sessionManager) Create(userId string) (string, error) {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
		UserID: userId,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

func (s *sessionManager) GetUserId(c echo.Context) (string, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return "", errors.New("JWT token missing or invalid")
	}

	if claims, ok := token.Claims.(*UserClaims); ok {
		return claims.UserID, nil
	}

	fmt.Printf("%T", token.Claims)
	return "", errors.New("failed to cast claims as UserClaims")

}

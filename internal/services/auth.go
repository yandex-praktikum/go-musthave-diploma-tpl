package services

import (
	"time"

	"github.com/KalessinD/gophermart/internal/common"
	"github.com/KalessinD/gophermart/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type (
	/*
		Объект службы для для генерации JWT
	*/
	AuthService struct {
		secretKey []byte
	}

	/*
		Интерфейс оОъекта службы для генерации JWT
	*/
	AuthInterface interface {
		GenerateToken(user *models.User, expireAt time.Time) (string, error)
	}
)

/*
Конструктор службы для генерации JWT.
*/
func NewAuthService(key string) AuthInterface {
	return &AuthService{
		secretKey: []byte(key),
	}
}

/*
Генерирует JWT строку
*/
func (s *AuthService) GenerateToken(user *models.User, expireAt time.Time) (string, error) {
	claims := &common.Claims{
		UserID: user.ID,
		Login:  user.Login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

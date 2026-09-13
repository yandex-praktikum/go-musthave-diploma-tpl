// Package utils содержит вспомогательные утилиты: проверку номеров заказов
// по алгоритму Луна и работу с JWT-токенами.
package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTUtils — утилиты для работы с JWT-токенами: генерация, проверка,
// обновление и извлечение данных пользователя.
type JWTUtils struct {
	// secretKey — секретный ключ для подписи и проверки токенов (HS256).
	secretKey []byte
}

// NewJWTUtils создаёт новый экземпляр JWTUtils с указанным секретным ключом.
func NewJWTUtils(secretKey string) *JWTUtils {
	return &JWTUtils{
		secretKey: []byte(secretKey),
	}
}

// Claims — структура полезной нагрузки JWT-токена.
//
// Помимо стандартных полей RegisteredClaims содержит идентификатор
// аутентифицированного пользователя.
type Claims struct {
	// UserID — идентификатор пользователя, для которого выдан токен.
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken генерирует JWT-токен для пользователя с идентификатором userID.
//
// Токен подписывается алгоритмом HS256 и имеет срок действия 24 часа.
// Возвращает строку токена или ошибку подписи.
func (u *JWTUtils) GenerateToken(userID int64) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(u.secretKey)
}

// ValidateToken проверяет подпись и срок действия JWT-токена.
//
// Поддерживается только алгоритм подписи HMAC (HS256). Возвращает разобранные
// claims или ошибку, если токен невалиден, просрочен или подписан
// неожиданным алгоритмом.
func (u *JWTUtils) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return u.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken продлевает срок действия токена.
//
// Проверяет переданный токен и, если он валиден, выпускает новый токен
// для того же пользователя с обновлённым сроком действия.
// Возвращает новый токен или ошибку валидации исходного.
func (u *JWTUtils) RefreshToken(tokenString string) (string, error) {
	claims, err := u.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Создаем новый токен с обновленным сроком
	return u.GenerateToken(claims.UserID)
}

// ExtractUserID извлекает идентификатор пользователя из валидного JWT-токена.
//
// Возвращает userID или ошибку, если токен невалиден.
func (u *JWTUtils) ExtractUserID(tokenString string) (int64, error) {
	claims, err := u.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

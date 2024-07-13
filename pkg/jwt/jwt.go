package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// Claims — структура утверждений, которая включает стандартные утверждения и
// одно пользовательское UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

type JWTClient struct {
	tokenExp  time.Duration
	secretKey string
}

var (
	ErrTokenInvalid            = errors.New("токен невалидный")
	ErrUnexpectedSigningMethod = errors.New("метод подписи токена невалидный")
)

func NewJwt(tokenExp time.Duration, secretKey string) *JWTClient {
	return &JWTClient{
		tokenExp:  tokenExp,
		secretKey: secretKey,
	}
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func (jwtClient *JWTClient) BuildJWTString(userID int) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtClient.tokenExp)),
		},
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(jwtClient.secretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func (jwtClient *JWTClient) GetUserID(tokenString string) (int, error) {
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnexpectedSigningMethod
		}

		return []byte(jwtClient.secretKey), nil
	})
	if err != nil {
		return -1, err
	}

	if !token.Valid {
		return -1, ErrTokenInvalid
	}

	// возвращаем ID пользователя в читаемом виде
	return claims.UserID, nil
}

func (jwtClient *JWTClient) GetTokenExp() time.Duration {
	return jwtClient.tokenExp
}

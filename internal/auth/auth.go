package auth

import (
	"fmt"
	//"log"
	"time"

	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/kindenko/gophermart/internal/models"
	"github.com/labstack/echo/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserLogin string
}

const TokenMaxAge = time.Hour * 3
const SecretKey = "very2secret3token"

func UserAuthorization() {

}

// func GenerateTokensAndSetCookies(user models.User, c echo.Context) error {
// 	accessToken, err := GenerateToken(user)
// 	if err != nil {
// 		return err
// 	}

// 	setTokenCookie("access-token", accessToken, c)
// 	setUserCookie(user, c)

// 	return nil
// }

func GenerateToken(user models.User) (string, error) {

	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenMaxAge)),
		},
		// собственное утверждение
		UserLogin: user.Login,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}
	// возвращаем строку токена
	return tokenString, nil
}

func ValidationToken(tokenString string) bool {

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SecretKey), nil
		})
	if err != nil {
		return false
	}

	if !token.Valid {
		return false
	}

	return true
}

func setTokenCookie(name, token string, c echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = name
	cookie.Value = token
	cookie.Path = "/"
	cookie.HttpOnly = true

	c.SetCookie(cookie)
}

func setUserCookie(user *models.User, c echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = "user"
	cookie.Value = user.Login
	cookie.Path = "/"
	c.SetCookie(cookie)
}

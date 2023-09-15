package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

const TokenExp = time.Hour * 3
const SecretKey = "secret"
const CookieName = "token"

// func generateJWTString() (string, error) {
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
// 		},
// 		UserID: uuid.NewString(),
// 	})
// 	return token.SignedString([]byte(SecretKey))
// }

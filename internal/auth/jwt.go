package auth

import (
    "time"

    "github.com/golang-jwt/jwt/v4"
)

// JWTSecretKey - секретный ключ для JWT токенов
var JWTSecretKey = []byte("your-secret-key-change-in-production")

// Claims представляет JWT claims
type Claims struct {
    UserID int    `json:"user_id"`
    Login  string `json:"login"`
    jwt.RegisteredClaims
}

// CreateToken создает JWT токен для пользователя
func CreateToken(userID int, login string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour)
    
    claims := &Claims{
        UserID: userID,
        Login:  login,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Subject:   login,
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(JWTSecretKey)
}

// VerifyToken проверяет JWT токен
func VerifyToken(tokenString string) (*Claims, error) {
    claims := &Claims{}
    
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return JWTSecretKey, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if !token.Valid {
        return nil, jwt.ErrSignatureInvalid
    }
    
    return claims, nil
}

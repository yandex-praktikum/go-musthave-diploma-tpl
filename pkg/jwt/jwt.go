// Package jwt provides a JWT interface for auth operations.
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

// NewToken generates a new JWT token for the given user id, secret and token duration.
// Using HS256 method for signing tokens.
// Returns access token strings and error.
func NewToken(uid string, accessDuration time.Duration, secret string) (string, error) {

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = uid
	claims["exp"] = time.Now().Add(accessDuration).Unix()
	claims["iat"] = time.Now().Unix()
	claims["scope"] = "access"

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("Signed fail: %v", err.Error())
	}

	return tokenString, nil
}

// get token from string
func GetToken(token string, secret []byte) (*jwt.Token, error) {
	tokenData, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	return tokenData, nil
}

// GetClaimsVerified returns the claims of a verified token.
// It is used for token verification with checking the signature.
//
// Use *jwt.Token as input parameter
func GetClaimsVerified(token *jwt.Token) (jwt.MapClaims, error) {
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrInvalidToken
}

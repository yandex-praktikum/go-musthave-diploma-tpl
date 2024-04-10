package domain

import "github.com/golang-jwt/jwt/v4"

type TokenString string

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

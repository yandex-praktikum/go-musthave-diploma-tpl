package common

import "github.com/golang-jwt/jwt/v5"

type (
	Claims struct {
		jwt.RegisteredClaims
		UserID string
		Login  string
	}
)

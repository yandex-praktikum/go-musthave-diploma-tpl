package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	TokenExp   = time.Hour * 3
	SecretKey  = "123"
	CookieName = "token"
)

var (
	nonAuthHandles = []string{
		"/api/user/register",
		"/api/user/login",
	}
)

func IsAuthHandles(url string) bool {
	for _, h := range nonAuthHandles {
		if h == url {
			return true
		}
	}
	return false
}

type claims struct {
	jwt.RegisteredClaims
	UserID string
}

type UserIDKey struct{}

func UpdateContext(ctx context.Context, generatedUserID string) context.Context {
	return context.WithValue(ctx, UserIDKey{}, generatedUserID)
}

func FromContext(ctx context.Context) string {
	userID, ok := ctx.Value(UserIDKey{}).(string)
	if !ok {
		return ""
	}
	return userID
}

func buildJWTString(ctx context.Context, generatedUserID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: generatedUserID,
	})
	return token.SignedString([]byte(SecretKey))
}

func GetUserID(tokenValue string) (string, error) {
	c := &claims{}
	token, err := jwt.ParseWithClaims(tokenValue, c, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(SecretKey), nil
	})
	if err != nil {
		return "", fmt.Errorf("failed parsing: %w, token=%s", err, tokenValue)
	}

	if !token.Valid {
		return "", fmt.Errorf("broken token")
	}
	return c.UserID, nil
}

func SetTokenToResponse(ctx context.Context, rw http.ResponseWriter, userID string) error {
	token, err := buildJWTString(ctx, userID)
	if err != nil {
		return err
	}

	http.SetCookie(rw, &http.Cookie{
		Name:  CookieName,
		Value: token,
	})

	return nil
}

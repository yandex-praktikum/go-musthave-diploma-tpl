package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/KalessinD/gophermart/internal/common"
	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/services"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/mock/gomock"
)

// тестирует генерацию JWT
func TestCommonAction_GenerateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("token generation", func(t *testing.T) {
		user := &models.User{
			ID:    "123",
			Login: "testuser",
		}
		expireAt := time.Now().Add(time.Hour)
		secret := "test-secret"

		svc := services.NewAuthService(secret)
		tokenStr, err := svc.GenerateToken(user, expireAt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Парсим токен для проверки содержимого
		claims := &common.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil {
			t.Fatalf("failed to parse token: %v", err)
		}

		if !token.Valid {
			t.Error("token is not valid")
		}

		if claims.UserID != "123" {
			t.Errorf("expected userID 123, got %s", claims.UserID)
		}
		if claims.Login != "testuser" {
			t.Errorf("expected login testuser, got %s", claims.Login)
		}
	})
}

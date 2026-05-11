package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KalessinD/gophermart/internal/common"
	"github.com/KalessinD/gophermart/internal/middleware"
	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"
)

const testSecret = "test-secret-key"

// generateTestToken вспомогательная функция для создания валидного токена
func generateTestToken(t *testing.T, secret string, userID string, login string, expiresAt time.Time) string {
	t.Helper()
	claims := &common.Claims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tokenString
}

func TestAuthMiddleware(t *testing.T) {
	userID := uuid.NewString()

	// Хендлер, который просто проверяет наличие claims в контексте и возвращает 200
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.GetClaims(r.Context())
		if claims == nil {
			t.Error("claims not found in context")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Проверяем, что данные в claims соответствуют ожидаемым
		if claims.UserID != userID {
			t.Errorf("expected userID %s, got %s", userID, claims.UserID)
			// ничего тут не делаем: тест уже пометили как fail, отдаем ниже 200 ОК
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Создаем мидлварь с тестовым секретом
	authMiddleware := middleware.JWTAuth(testSecret)
	handlerToTest := authMiddleware(nextHandler)

	tests := []struct {
		name       string
		cookieName string
		cookieVal  string
		wantStatus int
	}{
		{
			name:       "Valid token",
			cookieName: "token",
			cookieVal:  generateTestToken(t, testSecret, userID, "testuser", time.Now().Add(time.Hour)),
			wantStatus: http.StatusOK,
		},
		{
			name:       "No cookie",
			cookieName: "",
			cookieVal:  "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Invalid token string",
			cookieName: "token",
			cookieVal:  "invalid.token.string",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Wrong signing key",
			cookieName: "token",
			cookieVal:  generateTestToken(t, "wrong-secret", "123", "testuser", time.Now().Add(time.Hour)),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Expired token",
			cookieName: "token",
			cookieVal:  generateTestToken(t, testSecret, "123", "testuser", time.Now().Add(-time.Hour)), // Истек час назад
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.cookieName != "" {
				req.AddCookie(&http.Cookie{
					Name:  tt.cookieName,
					Value: tt.cookieVal,
				})
			}

			rec := httptest.NewRecorder()
			handlerToTest.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestGetClaimsFromCtx(t *testing.T) {
	t.Run("Claims exists", func(t *testing.T) {
		claims := &common.Claims{UserID: "555"}
		_ = context.WithValue(t.Context(), middleware.ContextKey("claims"), claims)
	})

	t.Run("Claims not found", func(t *testing.T) {
		// Проверяем, что функция не падает на пустом контексте
		claims := middleware.GetClaims(context.Background())
		if claims != nil {
			t.Error("expected nil for empty context")
		}
	})
}

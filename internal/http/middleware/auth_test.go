package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/anon-d/gophermarket/pkg/jwt"
)

const testSecret = "test-secret-key"

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthMiddleware_ValidTokenInHeader(t *testing.T) {
	token, _ := jwt.NewToken("user-123", time.Hour, testSecret)

	router := gin.New()
	router.Use(AuthMiddleware(testSecret))
	router.GET("/test", func(c *gin.Context) {
		userID := GetUserID(c)
		c.String(http.StatusOK, userID)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Body.String() != "user-123" {
		t.Errorf("Expected user-123, got %s", w.Body.String())
	}
}

func TestAuthMiddleware_ValidTokenInCookie(t *testing.T) {
	token, _ := jwt.NewToken("user-456", time.Hour, testSecret)

	router := gin.New()
	router.Use(AuthMiddleware(testSecret))
	router.GET("/test", func(c *gin.Context) {
		userID := GetUserID(c)
		c.String(http.StatusOK, userID)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Body.String() != "user-456" {
		t.Errorf("Expected user-456, got %s", w.Body.String())
	}
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware(testSecret))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware(testSecret))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	token, _ := jwt.NewToken("user-123", time.Hour, "different-secret")

	router := gin.New()
	router.Use(AuthMiddleware(testSecret))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	token, _ := jwt.NewToken("user-123", -time.Hour, testSecret)

	router := gin.New()
	router.Use(AuthMiddleware(testSecret))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_MalformedAuthHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "token-without-bearer"},
		{"wrong prefix", "Basic sometoken"},
		{"empty bearer", "Bearer "},
		{"only bearer", "Bearer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(AuthMiddleware(testSecret))
			router.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "ok")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.header)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected status %d, got %d for header %q", http.StatusUnauthorized, w.Code, tt.header)
			}
		})
	}
}

func TestAuthMiddleware_CookiePriorityOverHeader(t *testing.T) {
	cookieToken, _ := jwt.NewToken("cookie-user", time.Hour, testSecret)
	headerToken, _ := jwt.NewToken("header-user", time.Hour, testSecret)

	router := gin.New()
	router.Use(AuthMiddleware(testSecret))
	router.GET("/test", func(c *gin.Context) {
		userID := GetUserID(c)
		c.String(http.StatusOK, userID)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: cookieToken})
	req.Header.Set("Authorization", "Bearer "+headerToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	// Cookie имеет приоритет
	if w.Body.String() != "cookie-user" {
		t.Errorf("Expected cookie-user, got %s", w.Body.String())
	}
}

func TestGetUserID_NoUserInContext(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		userID := GetUserID(c)
		c.String(http.StatusOK, userID)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Body.String() != "" {
		t.Errorf("Expected empty string, got %s", w.Body.String())
	}
}

func TestAuthMiddleware_BearerCaseInsensitive(t *testing.T) {
	token, _ := jwt.NewToken("user-123", time.Hour, testSecret)

	tests := []struct {
		name   string
		header string
	}{
		{"lowercase bearer", "bearer " + token},
		{"uppercase BEARER", "BEARER " + token},
		{"mixed case BeArEr", "BeArEr " + token},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(AuthMiddleware(testSecret))
			router.GET("/test", func(c *gin.Context) {
				userID := GetUserID(c)
				c.String(http.StatusOK, userID)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.header)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d for header %q", http.StatusOK, w.Code, tt.header)
			}
		})
	}
}

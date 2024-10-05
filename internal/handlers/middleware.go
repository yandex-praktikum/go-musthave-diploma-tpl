package handlers

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal/adapters"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal/utils"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id,omitempty"`
}

func WithLogging(logger zap.SugaredLogger, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		logger.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
}

type responseData struct {
	status int
	size   int
}

// добавляем реализацию http.ResponseWriter
type loggingResponseWriter struct {
	http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
	responseData        *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

const JwtTTL = 15 * time.Minute

func BuildJWTString(secretKey string, userID uuid.UUID) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(JwtTTL)),
		},
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func fetchUserIDFromToken(secretKey string, tokenStr string) (uuid.UUID, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
	if err != nil {
		return uuid.Nil, err
	}

	return claims.UserID, nil
}

func WithAuth(authRequired bool, h http.HandlerFunc) http.HandlerFunc {
	hostname := utils.Must(url.Parse(internal.Config.BaseURL)).Hostname()
	return func(w http.ResponseWriter, r *http.Request) {
		authCookie, _ := r.Cookie("access_token")

		var userID uuid.UUID
		var accessToken string

		if authCookie != nil && authCookie.Value != "" {
			accessToken = authCookie.Value
		}

		authHeader := r.Header.Get("Authorization")
		if accessToken == "" && authHeader != "" {
			accessToken = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if accessToken != "" {
			var err error
			userID, err = fetchUserIDFromToken(internal.Config.JwtSecret, accessToken)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("invalid access token"))
				return
			}
		}
		if userID == uuid.Nil {
			if authRequired {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("Unauthorized"))
				return
			}
			userID = uuid.New()
		}

		if accessToken == "" {
			var err error
			accessToken, err = BuildJWTString(internal.Config.JwtSecret, userID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			Path:     "/",
			Domain:   hostname,
			Expires:  time.Now().Add(JwtTTL),
			Secure:   true,
			HttpOnly: true,
		})

		h.ServeHTTP(w, r.WithContext(adapters.UserIDToCxt(r.Context(), userID)))
	}
}

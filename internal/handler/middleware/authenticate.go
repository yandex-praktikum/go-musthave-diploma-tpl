package middleware

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/AnxVit/go-musthave-diploma-tpl/internal/logger"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/utils"
)

var (
	ErrTokenExpired = errors.New("token has expired")
)

// middleware аутентификации
// берет токен из cookie, если не найден переадресовывает на /api/user/login
func Authenticate(encryptionKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "", http.StatusUnauthorized)
				http.Redirect(w, r, "/api/user/login", http.StatusMovedPermanently)
				logger.Log.Info("couldn't get token from cookie", zap.Error(err))
				return
			}
			tokenString := cookie.Value

			userID, err := authenticate(tokenString, encryptionKey)
			if err != nil {
				if errors.Is(err, ErrTokenExpired) {
					http.Error(w, "token expired", http.StatusUnauthorized)
					logger.Log.Info("token expired", zap.Any("user-id", userID))
					return
				}
				http.Error(w, "", http.StatusInternalServerError)
				logger.Log.Warn("Couldn't authenticate", zap.Error(err))
				return
			}

			ctx := r.Context()

			ctx = utils.SetUserID(ctx, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func authenticate(credentials, encryptionKey string) (string, error) {
	encrypted, err := jwt.ParseSigned(credentials)
	if err != nil {
		return "", errors.WithMessage(err, "parsing")
	}

	if _, ok := encrypted.Headers[0].ExtraHeaders[jose.HeaderContentType]; ok {
		return "", errors.New("filtered by header")
	}

	claims := jwt.Claims{}
	if err = encrypted.Claims([]byte(encryptionKey), &claims); err != nil {
		return "", errors.WithMessage(err, "decryption")
	}
	if err = claims.Validate(jwt.Expected{Time: time.Now()}); err != nil {
		return "", ErrTokenExpired
	}

	return claims.Subject, nil
}

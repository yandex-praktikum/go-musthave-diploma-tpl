package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
)

type key string

const keyUserId key = "userID"

func NewContext(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, keyUserId, userID)
}

func FromContext(ctx context.Context) uuid.UUID {
	return ctx.Value(keyUserId).(uuid.UUID)
}

func Authenticate(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID uuid.UUID
			switch r.URL.Path {
			case "/api/user/register":
			case "/api/user/login":
			default:
				authToken := r.Header.Get("Authorization")

				if len(authToken) == 0 {
					writeResponse(w, http.StatusUnauthorized, Error{Message: `No "Authorization" header provided`})
					return
				}

				h := strings.Split(authToken, " ")
				if len(h) != 2 {
					writeResponse(w, http.StatusUnauthorized, Error{Message: "Incorrect auth header"})
					return
				}

				if strings.ToLower(h[0]) != "bearer" {
					writeResponse(w, http.StatusUnauthorized, Error{Message: "Incorrect auth header"})
					return
				}

				token, err := jwt.ParseString(h[1], jwt.WithVerify(jwa.HS256, secret), jwt.WithValidate(true))
				if err != nil {
					writeResponse(w, http.StatusUnauthorized, Error{Message: "Token verification error"})
					return
				}

				userID, err = uuid.Parse(token.Subject())
				if err != nil {
					writeResponse(w, http.StatusUnauthorized, Error{Message: "Token parsing error"})
					return
				}
			}
			next.ServeHTTP(w, r.WithContext(NewContext(r.Context(), userID)))
		})
	}
}

func LogRequest(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			logger.Info("request", slog.Any("URL", r.URL.Path), slog.String("RemoteAddr", r.RemoteAddr), slog.Int("status", ww.Status()))
		})
	}
}

func writeResponse(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`message: "Internal server error`))
		return
	}
	w.WriteHeader(code)
	w.Write(b)
}

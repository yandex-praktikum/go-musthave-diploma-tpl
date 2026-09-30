package handler

import (
	"context"
	"mime"
	"net/http"
)

type ctxKey int

const (
	usernameKey        ctxKey = 0
	maxRequestBodySize        = 1 << 20
)

func requireJSONContent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mediatype, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		if err != nil || mediatype != "application/json" {
			http.Error(w, "application/json only", http.StatusBadRequest)
			return
		}
		req.Body = http.MaxBytesReader(w, req.Body, maxRequestBodySize)
		next.ServeHTTP(w, req)
	})
}

func (h *apiHandler) validateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "user is not authorized", http.StatusUnauthorized)
			return
		}

		username, err := h.svc.IdentifyUser(authHeader)
		if err != nil {
			h.logger.Debug(err.Error())
			http.Error(w, "user is not authorized", http.StatusUnauthorized)
			return
		}

		if username == "" {
			h.logger.Debug("IdentifyUser returned an empty username for a valid token")
			http.Error(w, "user is not authorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(req.Context(), usernameKey, username)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

// usernameFrom returns the username stored in the context by validateToken, or "".
func usernameFromContext(ctx context.Context) string {
	username, _ := ctx.Value(usernameKey).(string)
	return username
}

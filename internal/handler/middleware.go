package handler

import (
	"context"
	"mime"
	"net/http"
	"strings"
)

// ctxKey is an unexported type for context keys; it prevents collisions
// with keys defined in other packages.
type ctxKey int

// usernameKey is the context key that identifies the authenticated username.
const usernameKey ctxKey = 0

// maxRequestBodySize limits the accepted request body size to 1 MiB.
const maxRequestBodySize = 1 << 20

// msgUnauthorized is written to the HTTP response when the request lacks a valid authorization token.
const msgUnauthorized = "user is not authorized"

// bearerPrefix is the required prefix of the Authorization header value.
const bearerPrefix = "Bearer "

// requireJSONContent rejects requests whose Content-Type is not application/json.
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

// validateToken rejects requests without a valid Bearer token in the Authorization
// header and stores the authenticated username in the request context.
func (h *apiHandler) validateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authHeader := req.Header.Get("Authorization")

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			http.Error(w, msgUnauthorized, http.StatusUnauthorized)
			return
		}

		username, err := h.svc.IdentifyUser(strings.TrimPrefix(authHeader, bearerPrefix))
		if err != nil {
			h.logger.Debug(err.Error())
			http.Error(w, msgUnauthorized, http.StatusUnauthorized)
			return
		}

		if username == "" {
			h.logger.Debug("IdentifyUser returned an empty username for a valid token")
			http.Error(w, msgUnauthorized, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(req.Context(), usernameKey, username)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

// usernameFromContext returns the username stored in the context by validateToken, or "".
func usernameFromContext(ctx context.Context) string {
	username, _ := ctx.Value(usernameKey).(string)
	return username
}

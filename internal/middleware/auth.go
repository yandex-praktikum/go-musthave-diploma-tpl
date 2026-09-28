// Package middleware contains HTTP middleware used by the service.
package middleware

import (
	"net/http"

	"github.com/true18/go-musthave-diploma-tpl/internal/auth"
)

// Auth protects routes by requiring a valid signed auth cookie.
func Auth(manager *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := manager.UserIDFromRequest(r)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(auth.WithUserID(r.Context(), userID)))
		})
	}
}

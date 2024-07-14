package logging

import (
	"context"
	"net/http"
)

type JWTClient interface {
	GetUserID(tokenString string) (int, error)
}

func WithPrivateRouter(h http.Handler, jwtClient JWTClient) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		atCookie, err := r.Cookie("at")
		if err != nil || atCookie == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		id, err := jwtClient.GetUserID(atCookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userId", id)

		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)

	}

	return fn
}

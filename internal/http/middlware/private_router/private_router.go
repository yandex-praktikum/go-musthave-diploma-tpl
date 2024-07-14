package privateRouter

import (
	"context"
	"net/http"
)

type UserExister interface {
	GetIsUserExistById(ctx context.Context, userId int) (bool, error)
}

type JWTClient interface {
	GetUserID(tokenString string) (int, error)
}

func WithPrivateRouter(h http.Handler, jwtClient JWTClient, userExister UserExister) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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

		isUserExist, err := userExister.GetIsUserExistById(ctx, id)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !isUserExist {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, "userId", id)

		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)

	}

	return fn
}

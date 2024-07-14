package privateRouter

import (
	"context"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/config"
	logging2 "github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	"net/http"
)

type UserExister interface {
	GetIsUserExistById(ctx context.Context, userId int) (bool, error)
}

type JWTClient interface {
	GetUserID(tokenString string) (int, error)
}

func WithPrivateRouter(h http.Handler, logger logging2.Logger, jwtClient JWTClient, userExister UserExister) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		atCookie, err := r.Cookie(config.AUTH_COOKIE)
		if err != nil {
			logger.Errorf("Ошибка получения куки авторизации %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if atCookie == nil {
			logger.Infof("нет куки %v", atCookie)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		id, err := jwtClient.GetUserID(atCookie.Value)
		if err != nil {
			logger.Errorf("Ошибка при получении id из токена %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		isUserExist, err := userExister.GetIsUserExistById(ctx, id)
		if err != nil {
			logger.Errorf("Произошла ошибка при получении пользователя с id %v БД: %v", id, err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !isUserExist {
			logger.Errorf("Пользователя с id %v нет БД", id)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, "userId", id)

		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)

	}

	return fn
}

package servers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/k-morozov/go-musthave-diploma-tpl/components/auth"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/logger"
)

func WithAuth() ServiceOption {
	return func(s *httpServer) {
		withAuth(s)
	}
}

func withAuth(s *httpServer) {

	s.engine.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

			var (
				userID string
				err    error
			)

			if !auth.IsAuthHandles(req.URL.String()) {
				if userID, err = getUserIDFromRequest(req); err != nil {
					rw.WriteHeader(http.StatusUnauthorized)
					return
				}
			}

			next.ServeHTTP(rw, req.WithContext(auth.UpdateContext(req.Context(), userID)))
		})
	})
}

func getUserIDFromRequest(req *http.Request) (string, error) {
	ctx := req.Context()
	logFromContext := logger.FromContext(ctx)
	lg := logFromContext.With().Caller().Logger()

	var resultUserID string

	if cookie, err := req.Cookie(auth.CookieName); errors.Is(err, http.ErrNoCookie) {
		lg.Info().
			Str("url", req.URL.String()).
			Msg("cookie not found on request")
		return "", fmt.Errorf("request without token")
	} else {
		if userID, err := auth.GetUserID(cookie.Value); err != nil {
			lg.Err(err).
				Msg("failed get user id from cookie. Create new.")
			return "", fmt.Errorf("request without token")
		} else {
			resultUserID = userID
			lg.Info().
				Str("userID", userID).
				Msg("req with cookie")
		}
	}
	lg.Info().Msg("auth done")
	return resultUserID, nil
}

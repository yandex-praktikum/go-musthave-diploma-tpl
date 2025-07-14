package routers

import (
	"net/http"

	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriterWithStatus{ResponseWriter: w, status: 200}
			next.ServeHTTP(rw, r)
			if rw.status == http.StatusInternalServerError {
				logger.Error("Внутренняя ошибка сервера", zap.String("url", r.URL.Path))
			}
		})
	}
}

type responseWriterWithStatus struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriterWithStatus) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

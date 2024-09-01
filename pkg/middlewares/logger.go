package middlewares

import (
	"net/http"
	"time"

	"github.com/eac0de/gophermart/pkg/logger"
	"go.uber.org/zap"
)

type (
	responseData struct {
		size   int
		status int
	}

	logResponseWriter struct {
		responseData *responseData
		http.ResponseWriter
	}
)

func (lw *logResponseWriter) Write(body []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(body)
	if err != nil {
		return 0, err
	}
	lw.responseData.size += size
	return size, err
}

func (lw *logResponseWriter) WriteHeader(statusCode int) {
	lw.ResponseWriter.WriteHeader(statusCode)
	lw.responseData.status = statusCode
}

func GetLoggerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var (
				respData = responseData{0, 0}
				lw       = logResponseWriter{responseData: &respData, ResponseWriter: w}
				duration time.Duration
			)
			start := time.Now()
			next.ServeHTTP(&lw, r)
			duration = time.Since(start)
			logger.Log.Info("HTTP request",
				zap.String("URI", r.URL.Path),
				zap.String("method", r.Method),
				zap.Duration("duration", duration),
				zap.Int("statusCode", lw.responseData.status),
				zap.Int("size", lw.responseData.size),
			)
		}
		return http.HandlerFunc(fn)
	}
}

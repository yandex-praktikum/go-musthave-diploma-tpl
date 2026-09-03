package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Logging writes method, URI, response status, size and duration to logger.
func Logging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(recorder, r)
			logger.Info("request",
				zap.String("method", r.Method),
				zap.String("uri", r.URL.RequestURI()),
				zap.Int("status", recorder.status),
				zap.Int("size", recorder.size),
				zap.Duration("duration", time.Since(started)),
			)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	size, err := r.ResponseWriter.Write(data)
	r.size += size
	return size, err
}

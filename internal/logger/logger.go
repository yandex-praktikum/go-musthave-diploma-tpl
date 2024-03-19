package logger

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

var Log *zap.Logger = zap.NewNop()

func Initialize(level, env string) error {
	logLevel, err := zap.ParseAtomicLevel(level)

	if err != nil {
		return err
	}

	var config zap.Config

	if env == "development" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	config.Level = logLevel

	logger, err := config.Build()

	if err != nil {
		return err
	}

	Log = logger

	return nil
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		wrappedWriter := newResponseWriter(w)

		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(startTime)

		Log.Info("Request processed",
			zap.String("URI", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", duration*time.Millisecond),
			zap.Int("status", wrappedWriter.statusCode),
		)
	})
}

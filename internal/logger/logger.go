package logger

import (
	"net/http"

	"go.uber.org/zap"
)

// Logger is the global zap logger; it is a no-op until InitLogger is called.
var Logger *zap.Logger = zap.NewNop()

// InitLogger creates a production zap logger at the given level.
func InitLogger(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zapLogger, err := cfg.Build()
	if err != nil {
		return err
	}

	Logger = zapLogger
	return nil
}

type (
	// responseInfo holds the HTTP status code and response body size.
	responseInfo struct {
		status int
		size   int
	}

	// loggingResponseWriter wraps http.ResponseWriter to capture status and size.
	loggingResponseWriter struct {
		http.ResponseWriter
		responseInfo *responseInfo
	}
)

// Write records the response body size and defaults status to 200.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseInfo.size += size

	if r.responseInfo.status == 0 {
		r.responseInfo.status = http.StatusOK
	}

	return size, err
}

// WriteHeader records the HTTP status code.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseInfo.status = statusCode
}

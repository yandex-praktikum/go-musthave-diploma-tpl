package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		gin.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		Log.Info("got incoming HTTP request",
			zap.String("URI", c.Request.RequestURI),
			zap.String("method", c.Request.Method),
			zap.Duration("duration", duration),
		)
	}
}

func ResponseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := &loggingResponseWriter{
			ResponseWriter: c.Writer,
			responseData:   responseData,
		}
		c.Writer = lw
		c.Next()

		Log.Info("response",
			zap.Int("status", lw.responseData.status),
			zap.Int("size", lw.responseData.size),
		)
	}
}

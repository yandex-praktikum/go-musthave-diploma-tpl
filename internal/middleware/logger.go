package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

const LoggerKey ContextKey = "logger"

// Extracts Logger from context
func GetLogger(ctx context.Context) *zap.Logger {
	if log, ok := ctx.Value(LoggerKey).(*zap.Logger); ok {
		return log
	}
	return nil
}

/*
Добавляет логгер в контекст
*/
func AddLoggerToContext(parentCtx context.Context, log *zap.Logger) context.Context {
	return context.WithValue(parentCtx, LoggerKey, log)
}

/*
Обёртка над http.ResponseWriter.Write
*/
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

/*
Обёртка над http.ResponseWriter.Writeheader
*/
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

/*
Из HTTP заголовков получает тип сжатия и добавляет его в zap-поле для логирования
*/
func GetEncodingField(prefix, name string, header http.Header) zap.Field {
	enc := header[name]
	if len(enc) == 0 {
		return zap.Skip()
	}
	return zap.Strings(prefix+strings.ToLower(name), enc)
}

/*
Мидлварь для добавления zap-логера с кастомными полями в контекст.

Добавляемые поля:
  - method - HTTP method
  - path - HTTP path
  - remote_addr - remote IP address
  - duration - время выполнения основной части запроса
  - status - HTTP status code
  - response_size - HTTP response size
  - request-content-encoding - HTTP заголовок Content-Encoding из запроса
  - request-accept-encoding - HTTP заголовок Accept-Encoding из запроса
  - responsecontent-encoding - HTTP заголовок Content-Encoding из ответа
  - response-accept-encoding - HTTP заголовок Accept-Encoding из ответа
*/
func Logger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			responseData := &responseData{
				status: 0,
				size:   0,
			}

			lw := &loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}

			extLogger := log.With(
				zap.String("request_id", middleware.GetReqID(r.Context())),
			)

			ctx := AddLoggerToContext(r.Context(), extLogger)

			next.ServeHTTP(lw, r.WithContext(ctx))

			extLogger.Info("request completed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Duration("duration", time.Since(start)),
				zap.Int("status", responseData.status),
				zap.Int("response_size", responseData.size),
				zap.Strings("request-content-encoding", r.Header.Values("Content-Encoding")),
				zap.Strings("request-accept-encoding", r.Header.Values("Accept-Encoding")),
				GetEncodingField("response-", "Content-Encoding", w.Header()),
				GetEncodingField("response-", "Accept-Encoding", w.Header()),
			)
		})
	}
}

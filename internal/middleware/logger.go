package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Для захвата данных ответа, которая иначе не доступна в мидлвеар
// получаем размер ответа в байт и статус код самого ответа
type responseData struct {
	size       int
	statusCode int
}

// Обертка над стандартным http.ResponseWriter, чтобы
// перехватывать вызовы Write и WriteHeader, затем сохранять данные (размер и статус)
// в responseData и делегировать вызовы оригинальному ResponseWriter
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (l *loggingResponseWriter) Write(p []byte) (int, error) {
	data, err := l.ResponseWriter.Write(p)
	l.responseData.size = data // здесь мы перехватываем размер ответа в байтах
	return data, err
}
func (l *loggingResponseWriter) WriteHeader(statusCode int) {
	l.ResponseWriter.WriteHeader(statusCode)
	l.responseData.statusCode = statusCode // записываем статус
}

func LoggingMiddleWare(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			responseData := responseData{0, 0}

			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   &responseData,
			}

			next.ServeHTTP(&lw, r)
			duration := time.Since(start)
			logger.Infoln(
				"uri", r.RequestURI,
				"method", r.Method,
				"status", responseData.statusCode,
				"size", responseData.size,
				"duration", duration,
			)
		})
	}
}

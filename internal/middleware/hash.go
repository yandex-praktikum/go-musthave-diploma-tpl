package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/IvanDolgov/go-musthave-diploma-tpl/internal/hash"
	"github.com/IvanDolgov/go-musthave-diploma-tpl/internal/logger"
)

// HashValidation middleware проверяет хеш входящих запросов
func HashValidation(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key != "" {
				// Читаем тело запроса
				body, err := io.ReadAll(r.Body)
				if err != nil {
					logger.Log.Error("Failed to read request body for hash validation")
					http.Error(w, "Bad request", http.StatusBadRequest)
					return
				}

				// Восстанавливаем тело для дальнейшей обработки
				r.Body = io.NopCloser(bytes.NewBuffer(body))

				// Получаем хеш из заголовка
				receivedHash := r.Header.Get("HashSHA256")

				// Проверяем хеш
				if !hash.VerifyHMACSHA256(body, receivedHash, key) {
					logger.Log.Warn("Hash validation failed")
					http.Error(w, "Bad request", http.StatusBadRequest)
					return
				}

				logger.Log.Debug("Hash validation successful")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// HashResponse middleware добавляет хеш к исходящим ответам
func HashResponse(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Используем custom ResponseWriter для перехвата ответа
			hw := &hashResponseWriter{ResponseWriter: w, key: key}
			next.ServeHTTP(hw, r)

			// Вычисляем хеш и добавляем в заголовок
			if hw.buffer.Len() > 0 {
				hashValue := hash.ComputeHMACSHA256(hw.buffer.Bytes(), key)
				if hashValue != "" {
					w.Header().Set("HashSHA256", hashValue)
				}
			}
		})
	}
}

// hashResponseWriter перехватывает запись ответа для вычисления хеша
type hashResponseWriter struct {
	http.ResponseWriter
	key    string
	buffer bytes.Buffer
}

func (hw *hashResponseWriter) Write(b []byte) (int, error) {
	// Сохраняем данные для вычисления хеша
	hw.buffer.Write(b)
	return hw.ResponseWriter.Write(b)
}

// Важно: перехватываем WriteHeader чтобы успеть вычислить хеш до отправки заголовков
func (hw *hashResponseWriter) WriteHeader(statusCode int) {
	// Вычисляем хеш перед отправкой заголовков
	if hw.buffer.Len() > 0 {
		hashValue := hash.ComputeHMACSHA256(hw.buffer.Bytes(), hw.key)
		if hashValue != "" {
			hw.ResponseWriter.Header().Set("HashSHA256", hashValue)
		}
	}
	hw.ResponseWriter.WriteHeader(statusCode)
}

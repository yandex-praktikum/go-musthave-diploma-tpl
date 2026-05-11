package server

import (
	"compress/gzip"
	"net/http"
	"strings"
	"sync"
)

// gzipPool пул gzip.Writer для переиспользования
var gzipPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

// Gzip middleware для сжатия ответов gzip
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, поддерживает ли клиент gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Создаём gzip writer из пула
		gz := gzipPool.Get().(*gzip.Writer)
		defer gzipPool.Put(gz)

		// Настраиваем writer для записи в response
		gz.Reset(w)
		defer gz.Close()

		// Устанавливаем заголовок Content-Encoding
		w.Header().Set("Content-Encoding", "gzip")

		// Оборачиваем response writer
		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			Writer:         gz,
		}

		next.ServeHTTP(gzw, r)
	})
}

// gzipResponseWriter обёртка для записи сжатых данных
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// GzipRequest middleware для распаковки gzip запросов
func GzipRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем Content-Encoding
		if r.Header.Get("Content-Encoding") != "gzip" {
			next.ServeHTTP(w, r)
			return
		}

		// Создаём gzip reader
		gzr, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		defer gzr.Close()

		// Заменяем тело запроса
		r.Body = gzr

		next.ServeHTTP(w, r)
	})
}

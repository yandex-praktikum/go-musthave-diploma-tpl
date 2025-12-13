package middleware

import (
	"net/http"
	"strings"

	"github.com/IvanDolgov/go-musthave-diploma-tpl/internal/compress"
)

// withGzip добавляет сжатие
func WithGzip(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		// Всегда распаковываем входящие сжатые данные
		if r.Header.Get("Content-Encoding") == "gzip" {
			cr, err := compress.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		// Всегда сжимаем исходящие данные если клиент поддерживает
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw := compress.NewCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		h.ServeHTTP(ow, r)
	})
}

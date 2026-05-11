package middleware

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strconv"
	"strings"
)

type (
	GzipWriter struct {
		http.ResponseWriter
		writer    *gzip.Writer
		threshold int
		buffer    *bytes.Buffer
		started   bool
		status    int
		headers   bool
	}
)

func Compression(threshold int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Распаковка входящего запроса
			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				gz, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "Failed to read gzip body", http.StatusBadRequest)
					return
				}
				defer gz.Close()
				r.Body = gz
			}

			// Проверка поддержки gzip клиентом
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				gw := NewGzipWriter(w, threshold)
				defer gw.Close()
				next.ServeHTTP(gw, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func NewGzipWriter(w http.ResponseWriter, threshold int) *GzipWriter {
	return &GzipWriter{
		ResponseWriter: w,
		threshold:      threshold,
		buffer:         &bytes.Buffer{},
		status:         http.StatusOK,
	}
}

func (g *GzipWriter) Write(b []byte) (int, error) {
	if g.started {
		return g.writer.Write(b)
	}

	n, err := g.buffer.Write(b)
	if err != nil {
		return n, err
	}

	// Если буфер превысил порог — включаем сжатие
	if g.buffer.Len() > g.threshold {
		g.startGzip()
	}

	return n, nil
}

func (g *GzipWriter) WriteHeader(statusCode int) {
	if g.headers {
		return
	}
	g.status = statusCode

	// Проверяем Content-Length, если хендлер его установил
	if cl := g.Header().Get("Content-Length"); cl != "" {
		if size, err := strconv.Atoi(cl); err == nil {
			if size > g.threshold {
				// Размер известен и больше порога — включаем сжатие сразу
				g.startGzip()
				return
			}
			// Размер известен и меньше порога — отправляем заголовки сразу без сжатия
			g.headers = true
			g.ResponseWriter.WriteHeader(g.status)
		}
	}
	// Если Content-Length нет или он некорректен, откладываем решение до Write
}

func (g *GzipWriter) startGzip() {
	if g.started {
		return
	}
	g.started = true

	// Удаляем Content-Length, так как размер изменится
	g.Header().Del("Content-Length")
	g.Header().Set("Content-Encoding", "gzip")

	if !g.headers {
		g.ResponseWriter.WriteHeader(g.status)
		g.headers = true
	}

	g.writer = gzip.NewWriter(g.ResponseWriter)

	// Сбрасываем буфер в gzip writer
	if g.buffer.Len() > 0 {
		_, _ = g.writer.Write(g.buffer.Bytes())
		g.buffer.Reset()
	}
}

func (g *GzipWriter) Close() error {
	if g.started {
		return g.writer.Close()
	}

	// Если сжатие не началось, просто отдаем буфер
	if !g.headers {
		g.ResponseWriter.WriteHeader(g.status)
	}

	// Устанавливаем реальный Content-Length для маленьких ответов
	if g.buffer.Len() > 0 {
		g.Header().Set("Content-Length", strconv.Itoa(g.buffer.Len()))
		_, _ = g.ResponseWriter.Write(g.buffer.Bytes())
	}
	return nil
}

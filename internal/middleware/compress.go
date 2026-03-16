package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type CompressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	if c.w.Header().Get("Content-Encoding") == "" {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	return c.zw.Write(p)
}

func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	c.Header().Set("Content-Encoding", "gzip")
	c.w.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	return c.zw.Close()
}

type CompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {

	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &CompressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *CompressReader) Read(p []byte) (int, error) {
	return c.zr.Read(p)
}

func (c *CompressReader) Close() error {
	err := c.r.Close()
	if err != nil {
		return err
	}
	return c.zr.Close()
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Распаковка данных
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := NewCompressReader(r.Body)
			if err != nil {
				http.Error(w, "failed to decompress gzip body", http.StatusBadRequest)
				return
			}
			defer cr.Close()
			r.Body = cr
		}

		// Подготовка сжатия исходных данных
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw := NewCompressWriter(w)
			defer cw.Close()
			w = cw
		}
		next.ServeHTTP(w, r)
	})
}

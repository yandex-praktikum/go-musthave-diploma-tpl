package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	Writer  http.ResponseWriter
	ZWriter *gzip.Writer
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	return w.ZWriter.Write(b)
}

func (w *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		w.Writer.Header().Set("Content-Encoding", "gzip")
	}
	w.Writer.WriteHeader(statusCode)
}

func (w *gzipWriter) Header() http.Header {
	return w.Writer.Header()
}

func (w *gzipWriter) Close() {
	w.ZWriter.Close()
}

func NewGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		Writer:  w,
		ZWriter: gzip.NewWriter(w),
	}
}

type gzipReader struct {
	Reader  io.ReadCloser
	ZReader *gzip.Reader
}

func (r *gzipReader) Close() error {
	if err := r.Reader.Close(); err != nil {
		return err
	}
	return r.ZReader.Close()
}

func (r *gzipReader) Read(p []byte) (n int, err error) {
	return r.ZReader.Read(p)
}

func newGzipReader(reader io.ReadCloser) (*gzipReader, error) {
	newReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		Reader:  reader,
		ZReader: newReader,
	}, nil
}

func (m *MiddleWare) GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentWriter := w

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			newWriter := NewGzipWriter(w)
			currentWriter = newWriter
			defer newWriter.Close()
		}

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			reader, err := newGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = reader
			defer reader.Close()
		}

		next.ServeHTTP(currentWriter, r)
	})
}

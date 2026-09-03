package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// Gzip decompresses gzip requests and compresses JSON responses for gzip clients.
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hasGzip(r.Header.Get("Content-Encoding")) {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer reader.Close()
			r.Body = readCloser{Reader: reader, Closer: r.Body}
		}

		if !hasGzip(r.Header.Get("Accept-Encoding")) {
			next.ServeHTTP(w, r)
			return
		}

		responseWriter := &gzipResponseWriter{ResponseWriter: w}
		next.ServeHTTP(responseWriter, r)
		if responseWriter.writer != nil {
			_ = responseWriter.writer.Close()
		}
	})
}

type readCloser struct {
	io.Reader
	io.Closer
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer      *gzip.Writer
	gzipEnabled bool
	wroteHeader bool
	status      int
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.status = status
	if isJSON(w.Header().Get("Content-Type")) && status != http.StatusNoContent {
		w.gzipEnabled = true
		w.writer = gzip.NewWriter(w.ResponseWriter)
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Del("Content-Length")
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if w.gzipEnabled {
		return w.writer.Write(data)
	}
	return w.ResponseWriter.Write(data)
}

func isJSON(contentType string) bool {
	return strings.Contains(contentType, "application/json")
}

func hasGzip(value string) bool {
	return strings.Contains(strings.ToLower(value), "gzip")
}

package logging

import (
	logging2 "github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	"github.com/sirupsen/logrus"
	"net/http"
	"runtime/debug"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	body        []byte
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body = b

	return rw.ResponseWriter.Write(b)
}

func WithLogging(h http.Handler, logger logging2.Logger) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logger.Error(
					"err", err,
					"trace", debug.Stack(),
				)
			}
		}()

		start := time.Now()
		wrapped := wrapResponseWriter(w)
		h.ServeHTTP(wrapped, r)

		logger.WithFields(logrus.Fields{
			"status":   wrapped.status,
			"path":     r.URL.EscapedPath(),
			"method":   r.Method,
			"duration": time.Since(start),
			"length":   r.ContentLength,
		}).Info("request completed")

	}

	return http.HandlerFunc(fn)
}

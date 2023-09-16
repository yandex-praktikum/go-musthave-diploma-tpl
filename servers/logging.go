package servers

import (
	"net/http"
	"strconv"
	"time"

	appLog "github.com/k-morozov/go-musthave-diploma-tpl/components/logger"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type (
	responseData struct {
		status int
		size   int
		header http.Header
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		data *responseData
	}
)

var _ http.ResponseWriter = &loggingResponseWriter{}

func OptLogger(lg zerolog.Logger) ServiceOption {
	return func(s *httpServer) {
		withLogger(s, lg)
	}
}

func (r *loggingResponseWriter) Header() http.Header {
	r.data.header = r.ResponseWriter.Header()
	return r.data.header
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.data.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(status int) {
	r.ResponseWriter.WriteHeader(status)
	r.data.status = status
}

func withLogger(s *httpServer, logger zerolog.Logger) {
	s.engine.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			uri := req.RequestURI
			method := req.Method
			headers := req.Header

			requestID := req.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
				req.Header.Set("X-Request-ID", requestID)
			}

			requestLogger := logger.With().Str("X-Request-ID", requestID).Logger()
			ctx := appLog.UpdateContext(req.Context(), requestLogger)

			requestLogger.Info().
				Str("URI", uri).
				Str("Method", method).
				Any("Headers", headers).
				Msg("request")

			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	})

	s.engine.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			start := time.Now()

			data := &responseData{}
			lgWriter := &loggingResponseWriter{
				rw,
				data,
			}

			next.ServeHTTP(lgWriter, req)

			duration := time.Since(start)

			requestID := req.Header.Get("X-Request-ID")

			logger.Info().
				Str("status", strconv.Itoa(data.status)).
				Str("X-Request-ID", requestID).
				Str("duration", duration.String()).
				Str("size", strconv.Itoa(data.size)).
				Any("Headers", data.header).
				Msg("response")
		})
	})
}

package middlewares

import (
	"encoding/base64"
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type Logging struct {
	log     *zap.Logger
	skip    map[string]struct{}
	logBody uint8
}

func NewLogging(log *zap.Logger, logBody uint8) *Logging {
	return &Logging{
		log: log, skip: map[string]struct{}{
			"/ping": {},
		},
		logBody: logBody,
	}
}

func (l *Logging) LoggingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		raw := r.URL.RawQuery

		if _, ok := l.skip[path]; ok {
			next.ServeHTTP(w, r)

			return
		}

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			if raw != "" {
				path = path + "?" + raw
			}

			lvl := zap.InfoLevel
			if ww.Status() > 399 {
				lvl = zap.ErrorLevel
			}

			var body []byte
			if l.logBody > 0 && r.PostForm != nil {
				if (l.logBody == 1 && ww.Status() > 399) || l.logBody > 1 {
					if b, err := json.Marshal(r.PostForm); err == nil {
						body = b
					}
				}
			}

			addr := r.RemoteAddr
			if forwarded := r.Header.Get("x-forwarded-for"); forwarded != "" {
				ip := strings.TrimSpace(strings.Split(forwarded, ",")[0])
				if ip != "" {
					addr = ip
				}
			}

			endTime := time.Now()
			fields := []zap.Field{
				zap.String("Url", path),
				zap.Int("StatusCode", ww.Status()),
				zap.Duration("Latency", endTime.Sub(start)),
				zap.String("ClientIp", addr),
				zap.Int("BodySize", ww.BytesWritten()),
			}

			if ww.Status() > 399 {
				httpRequest, _ := httputil.DumpRequest(r, false)
				fields = append(fields, zap.String("request", string(httpRequest)))
			}

			if len(body) > 0 {
				fields = append(fields, zap.String("Body", base64.StdEncoding.EncodeToString(body)))
			}

			if clientId := r.Header.Get("x-client-id"); clientId != "" {
				fields = append(fields, zap.String("client_id", clientId))
			}

			l.log.Log(lvl, r.Method, fields...)
		}()

		next.ServeHTTP(ww, r)
	}

	return http.HandlerFunc(fn)
}

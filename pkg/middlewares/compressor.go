package middlewares

import (
	"net/http"
	"strings"

	"github.com/eac0de/gophermart/pkg/compressor"
)

func GetGzipMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ow := w
			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")

			if supportsGzip {
				cw := compressor.NewCompressWriter(w,)
				ow = cw
				defer cw.Close()
			}
			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := compressor.NewCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer cr.Close()
			}
			next.ServeHTTP(ow, r)
		}
		return http.HandlerFunc(fn)
	}
}

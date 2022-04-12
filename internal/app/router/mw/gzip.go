package mw

import (
	"compress/gzip"
	resp "github.com/EestiChameleon/GOphermart/internal/app/router/responses"
	"io"
	"log"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

// ResponseGZIP - middleware that provides http.ResponseWriter with gzip Writer
// when header Accept-Encoding contains gzip and set w.header Content-Encoding = gzip
func ResponseGZIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get(resp.HeaderAcceptEncoding), "gzip") {
			log.Println("request accept-encoding != gzip => next.ServeHTTP")
			next.ServeHTTP(w, r)
			return
		}
		log.Println("request accept-encoding contains gzip")
		// создаём gzip.Writer поверх текущего w
		gz, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)
		if err != nil {
			log.Println("ResponseGZIP: gzip.NewWriterLevel(w, gzip.DefaultCompression) err: ", err)
			resp.NoContent(w, http.StatusBadRequest)
			return
		}
		defer gz.Close()
		log.Println("ResponseGZIP: response header set -> Content-encoding = gzip")
		w.Header().Set(resp.HeaderContentEncoding, "gzip")
		log.Println("ResponseGZIP: next.ServeHTTP")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

// RequestGZIP - middleware that decompress request.body when header Content-Encoding = gzip
func RequestGZIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			log.Println("RequestGZIP mw: Content-Encoding = gzip. Start to replace r.Body")
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				log.Println("RequestGZIP mw: gzip.NewReader(r.Body) err", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			log.Println("RequestGZIP mw: r.Body = gz")
			r.Body = gz
		}
		log.Println("RequestGZIP mw: next.ServeHTTP")
		next.ServeHTTP(w, r)
	})
}

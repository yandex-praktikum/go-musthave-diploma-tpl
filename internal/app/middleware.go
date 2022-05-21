package app

import (
	"compress/gzip"
	"net/http"
)

type gzipBodyReader struct {
	gzipReader *gzip.Reader
}

func (gz *gzipBodyReader) Read(p []byte) (n int, err error) {
	return gz.gzipReader.Read(p)
}

func (gz *gzipBodyReader) Close() error {
	return gz.gzipReader.Close()
}

func DecompressGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			r.Body = &gzipBodyReader{gzipReader: gz}
		}
		next.ServeHTTP(w, r)
	})
}

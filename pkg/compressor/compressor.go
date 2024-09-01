package compressor

import (
	"compress/gzip"
	"io"
	"net/http"
)

// Реализует http.ResponseWriter, нужен для сжатия и отправки сжатых данных
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	zw := gzip.NewWriter(w)
	return &compressWriter{
		w:  w,
		zw: zw,
	}
}

func (cw *compressWriter) Write(p []byte) (int, error) {
	return cw.zw.Write(p)
}

func (cw *compressWriter) Header() http.Header {
	return cw.w.Header()
}

func (cw *compressWriter) WriteHeader(statusCode int) {
	cw.w.Header().Set("Content-Encoding", "gzip")
	cw.w.WriteHeader(statusCode)
}

func (cw *compressWriter) Close() error {
	return cw.zw.Close()

}

// Реализует io.ReadCloser, нужен для чтения сжатых данных
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (cr compressReader) Read(p []byte) (n int, err error) {
	return cr.zr.Read(p)
}

func (cr *compressReader) Close() error {
	if err := cr.r.Close(); err != nil {
		return err
	}
	return cr.zr.Close()
}

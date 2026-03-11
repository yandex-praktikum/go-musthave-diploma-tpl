package gzip

import (
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	gzipcompress "compress/gzip"
)

const encodingGzip = "gzip"

// CompressWriter оборачивает gin.ResponseWriter и сжимает запись через gzip.
type CompressWriter struct {
	gin.ResponseWriter
	Writer *gzipcompress.Writer
}

// Write пишет сжатые данные в нижележащий Writer.
func (c *CompressWriter) Write(data []byte) (int, error) {
	return c.Writer.Write(data)
}

// WriteString пишет строку в сжатом виде.
func (c *CompressWriter) WriteString(s string) (int, error) {
	return c.Writer.Write([]byte(s))
}

// Close закрывает gzip.Writer (дозаписывает gzip-футер). Вызывать через defer в middleware.
func (c *CompressWriter) Close() error {
	return c.Writer.Close()
}

// SupportsGzipResponse возвращает true, если заголовок Accept-Encoding допускает gzip для ответа.
func SupportsGzipResponse(acceptEncoding string) bool {
	return strings.Contains(acceptEncoding, encodingGzip)
}

// SupportsGzipRequest возвращает true, если заголовок Content-Encoding указывает на gzip в теле запроса.
func SupportsGzipRequest(contentEncoding string) bool {
	return strings.Contains(contentEncoding, encodingGzip)
}

// NewCompressWriter создаёт обёртку над c.Writer для сжатия ответа. После обработки запроса нужно вызвать Close() (например, через defer).
func NewCompressWriter(w gin.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		ResponseWriter: w,
		Writer:         gzipcompress.NewWriter(w),
	}
}

// WrapRequestBody подменяет тело запроса читателем, распаковывающим gzip. Если body не в gzip — возвращает body без изменений.
// Caller должен закрывать возвращённый ReadCloser при необходимости (например, reader от gzip.NewReader).
func WrapRequestBody(body io.ReadCloser) (io.ReadCloser, error) {
	if body == nil {
		return nil, nil
	}
	reader, err := gzipcompress.NewReader(body)
	if err != nil {
		_ = body.Close()
		return nil, err
	}
	return reader, nil
}

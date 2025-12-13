package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
)

// GzipCompress сжимает данные в gzip
func GzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	// Используем defer для гарантированного закрытия writer даже при ошибках
	defer gz.Close()

	if _, err := gz.Write(data); err != nil {
		return nil, err
	}

	// Явно закрываем writer для финализации данных
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	// Writers - группа связанная с записью данных
	w  http.ResponseWriter
	zw *gzip.Writer

	// State - группа связанная с состоянием writer
	wroteHeader bool
}

func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	// Если заголовок еще не отправлен, отправляем его с кодом 200
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if !c.wroteHeader {
		// Устанавливаем Content-Encoding ДО вызова WriteHeader у оригинального writer
		c.w.Header().Set("Content-Encoding", "gzip")
		c.w.WriteHeader(statusCode)
		c.wroteHeader = true
	}
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	// Readers - группа связанная с чтением данных
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

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	// Закрываем оба reader'а, порядок важен
	if err := c.zr.Close(); err != nil {
		c.r.Close() // все равно пытаемся закрыть оригинальный reader
		return err
	}
	return c.r.Close()
}

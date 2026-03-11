package middleware

import (
	"net/http"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/gzip"
	"github.com/gin-gonic/gin"
)

// Gzip обрабатывает сжатие ответов (Accept-Encoding: gzip) и распаковку тел запросов (Content-Encoding: gzip).
// Логика сжатия/распаковки вынесена в пакет internal/gzip.
func Gzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptEncoding := c.GetHeader("Accept-Encoding")
		if gzip.SupportsGzipResponse(acceptEncoding) {
			cw := gzip.NewCompressWriter(c.Writer)
			defer cw.Close()
			c.Writer = cw
			c.Header("Content-Encoding", "gzip")
		}

		contentEncoding := c.GetHeader("Content-Encoding")
		if gzip.SupportsGzipRequest(contentEncoding) {
			body, err := gzip.WrapRequestBody(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid gzip body"})
				return
			}
			defer body.Close()
			c.Request.Body = body
		}

		c.Next()
	}
}

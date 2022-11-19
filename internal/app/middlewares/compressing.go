package middlewares

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
)

// gzipBodyWriter записывает тело ответа, сжатое с помощью gzip
type gzipBodyWriter struct {
	gin.ResponseWriter
	gz *gzip.Writer
}

func (g *gzipBodyWriter) WriteString(s string) (int, error) {
	return g.gz.Write([]byte(s))
}

func (g *gzipBodyWriter) Write(data []byte) (int, error) {
	return g.gz.Write(data)
}

type gzipBodyReader struct {
	body io.ReadCloser
	gz   *gzip.Reader
}

func (g *gzipBodyReader) Read(p []byte) (n int, err error) {
	return g.gz.Read(p)
}

func (g *gzipBodyReader) Close() error {
	if err := g.gz.Close(); err != nil {
		return err
	}
	if err := g.body.Close(); err != nil {
		return err
	}
	return nil
}

// CompressingResponseMiddleware осуществляет сжатие ответов сервера в формате gzip
func CompressingResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// если клиент принимает сжатые в gzip ответы
		if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Interval server error"})
				return
			}
			// по завершении функции закрываем writer
			defer func() {
				err := gz.Close()
				if err != nil {
					return
				}
			}()

			// на место обычного ResponseWriter записываем writer с gzip сжатием
			c.Writer = &gzipBodyWriter{c.Writer, gz}
			c.Header("Content-Encoding", "gzip")
		}
		c.Next()
	}
}

// DecompressingRequestMiddleware осуществляет распаковку запросов, сжатых в формает gzip
func DecompressingRequestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// если запрос пришел в сжатом виде
		if c.Request.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Interval server error"})
			}

			// по завершении функции закрываем reader
			defer func() {
				err := gz.Close()
				if err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Interval server error"})
					return
				}
			}()

			// перезаписываем тело запроса
			c.Request.Body = &gzipBodyReader{body: c.Request.Body, gz: gz}
		}

		c.Next()
	}
}

package handler

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"musthave/internal/service"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Handlers struct {
	Market     *service.Market // внутриняя логика приложения
	httpServer *echo.Echo
	secret     string
}

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewHandlers(market *service.Market) *Handlers {
	return &Handlers{Market: market}
}

// StartHTTP - инициализация и запуск сервера
func (h *Handlers) StartHTTP(ctx context.Context, httpPort, sk string) error {
	h.secret = sk
	h.httpServer = echo.New()
	h.httpServer.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogLatency:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			fields := []any{
				"method", c.Request().Method,
				"uri", c.Path(),
				"status", v.Status,
				"latency_ms", v.Latency.Milliseconds(),
				"request_id", requestID,
			}

			if v.Error != nil {
				fields = append(fields, "error", v.Error.Error())
				slog.Error("request failed", fields...)
			} else {
				slog.Info("request completed", fields...)
			}
			return nil
		},
	})) //в билиотеке уже есть middleware для логирования запрсов
	h.httpServer.Use(middleware.Recover())
	h.httpServer.Use(GzipMiddleware)
	//h.httpServer.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
	//	return h.authMiddleware(next)
	//})
	h.httpServer.HTTPErrorHandler = func(err error, c echo.Context) {
		// Логируем ошибку
		c.Logger().Error(err)

		// Определяем код и сообщение
		code := http.StatusInternalServerError
		message := "Internal Server Error"

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			message = fmt.Sprintf("%v", he.Message)
		}

		// Отправляем JSON
		if !c.Response().Committed {
			c.JSON(code, map[string]string{
				"error": message,
			})
		}
	}
	h.httpServer.POST("/api/user/login", h.logIn)
	h.httpServer.POST("/api/user/register", h.regitsterUser)
	h.httpServer.GET("/api/user/balance", h.withAuth(h.getBalance))
	h.httpServer.GET("/api/user/orders", h.withAuth(h.getOrderList))
	h.httpServer.POST("/api/user/orders", h.withAuth(h.setOrder))
	h.httpServer.POST("/api/user/balance/withdraw", h.withAuth(h.withdrawPoints))
	h.httpServer.GET("/api/user/withdrawals", h.withAuth(h.infoWithdrawals))

	go func() {
		if err := h.httpServer.Start(httpPort); err != nil && err != http.ErrServerClosed {
			h.Market.Lg.Error("HTTP сервер завершился с ошибкой", "error", err)
		}
	}()
	<-ctx.Done()
	h.Market.Lg.Info("Контекст завершен")
	h.StopHTTP(ctx)
	return nil

}

// остнавка http сервера
func (h *Handlers) StopHTTP(ctx context.Context) {
	h.Market.Lg.Info("Остановка сервера")
	h.httpServer.Shutdown(ctx)
}
func GzipMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Проверяем, что клиент поддерживает сжатие ответа
		acceptEncoding := c.Request().Header.Get(echo.HeaderAcceptEncoding)
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		// Проверяем, что клиент отправил сжатые данные
		contentEncoding := c.Request().Header.Get(echo.HeaderContentEncoding)
		sendsGzip := strings.Contains(contentEncoding, "gzip")

		// Если клиент отправил сжатые данные, оборачиваем тело запроса
		if sendsGzip {
			cr, err := newCompressReader(c.Request().Body)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			c.Request().Body = cr
			defer cr.Close()
		}

		// Если клиент поддерживает сжатие ответа, оборачиваем ResponseWriter
		if supportsGzip {
			res := c.Response()
			cw := newCompressWriter(res.Writer)
			res.Writer = cw
			defer cw.Close()
			res.Header().Set(echo.HeaderContentEncoding, "gzip")
		}

		// Передаём управление следующему обработчику
		return next(c)
	}
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

//unc (h *Handlers) getBody(ctx echo.Context) ([]byte, error) {
//	body, err := io.ReadAll(ctx.Request().Body)
//	if err != nil {
//		h.Market.Lg.Error("ошибка при работе с телом запроса: " + err.Error())
//		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
//	}
//	return body, nil
//

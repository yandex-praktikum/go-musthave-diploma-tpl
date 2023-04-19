package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/slog"
)

func InitLogger() *slog.Logger {
	textHandler := slog.NewTextHandler(os.Stdout)
	logger := slog.New(textHandler)

	return logger
}

func MiddlewareLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		if err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("method = %s, patch = %s,  IP = %s, statusCode = %d, timeResponse=%v", c.Method(), c.Path(), c.IP(), c.Response().StatusCode(), time.Since(start)))
		return nil
	}
}

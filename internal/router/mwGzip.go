package router

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/labstack/echo"
)

func (s *serverMart) gzip(next echo.HandlerFunc) echo.HandlerFunc {
	fmt.Println("=> gzip run")

	return func(c echo.Context) error {
		if c.Request().Header.Get("Content-Encoding") != "gzip" {
			return next(c)
		}

		qzBody, err := gzip.NewReader(c.Request().Body)
		if err != nil {
			return fmt.Errorf("qz is not exist")
		}

		body, err := io.ReadAll(qzBody)
		if err != nil {
			c.Error(echo.ErrInternalServerError)
			return fmt.Errorf("URL does not exist")
		}
		stringReader := bytes.NewReader(body)

		c.Request().Body = io.NopCloser(stringReader)

		return next(c)
	}
}

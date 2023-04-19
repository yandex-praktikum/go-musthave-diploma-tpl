package http

import "github.com/gofiber/fiber/v2"

func ErrorResponse(err error) *fiber.Map {
	return &fiber.Map{
		"status": false,
		"error":  err.Error(),
	}
}

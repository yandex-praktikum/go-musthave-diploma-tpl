package server

import "github.com/Raime-34/gophermart.git/internal/dto"

//go:generate mockgen -source=cookie_handler_interface.go -destination=../../mocks/cookie_handler.go -package=mock
type cookieHandlerInterface interface {
	Set(string, *dto.UserData)
	Get(string) (*dto.UserData, bool)
}

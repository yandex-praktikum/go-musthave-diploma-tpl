package server

import "github.com/Raime-34/gophermart.git/internal/dto"

type cookieHandlerInterface interface {
	Set(string, *dto.UserData)
	Get(string) *dto.UserData
}

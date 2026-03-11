package handler

import "github.com/MaxRadzey/go-musthave-diploma-tpl/internal/service"

// Handler — HTTP-хендлеры.
type Handler struct {
	Services     *service.Services
	CookieSecret string
}

// New создаёт Handler с контейнером сервисов и секретом для куки.
func New(services *service.Services, cookieSecret string) *Handler {
	return &Handler{Services: services, CookieSecret: cookieSecret}
}

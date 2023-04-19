package http

import (
	"github.com/RedWood011/cmd/gophermart/internal/config"
	"github.com/RedWood011/cmd/gophermart/internal/database/postgres"
	"github.com/RedWood011/cmd/gophermart/internal/logger"
	"github.com/RedWood011/cmd/gophermart/internal/service"
	"github.com/RedWood011/cmd/gophermart/internal/transport/http/auth"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/slog"
)

type Controller struct {
	service *service.Service
	storage *postgres.Repository
	cfg     *config.Config
	log     *slog.Logger
}

type ServerParams struct {
	Service *service.Service
	Storage *postgres.Repository
	Cfg     *config.Config
	Logger  *slog.Logger
}

func NewController(params ServerParams) *Controller {
	return &Controller{
		service: params.Service,
		storage: params.Storage,
		cfg:     params.Cfg,
		log:     params.Logger,
	}
}

func (c *Controller) HelloWorld(ctx *fiber.Ctx) error {
	id := ctx.Locals("user_id")
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": id,
	})

}

func NewServer(params ServerParams) *fiber.App {
	controller := NewController(params)
	app := fiber.New()
	app.Use(logger.MiddlewareLogger(params.Logger))
	app.Post("/api/user/register", controller.UserRegistration())
	app.Post("/api/user/login", controller.UserAuthorization())
	app.Post("/api/user/orders", auth.MiddlewareJWT(controller.cfg.Token), controller.CreateOrder())
	app.Get("/api/user/orders", auth.MiddlewareJWT(controller.cfg.Token), controller.GetOrders())
	app.Get("/api/user/balance", auth.MiddlewareJWT(controller.cfg.Token), controller.GetBalance())
	app.Post("/api/user/balance/withdraw", auth.MiddlewareJWT(controller.cfg.Token), controller.CreateWithdraw())
	app.Get("/api/user/balance/withdrawals", auth.MiddlewareJWT(controller.cfg.Token), controller.GetWithdraws())
	return app
}

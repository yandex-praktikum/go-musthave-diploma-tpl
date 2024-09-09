package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/handlers/authorize"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/handlers/order"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/handlers/register"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/middleware"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/auth"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/storage/db"
	"net/http"
	"os"
)

func main() {
	// инициализируем Config
	cfg := NewConfig()
	cfg.Parsed()

	// инициализируем Logger
	logs := logger.NewLogger(logger.WithLevel(cfg.LogLevel))
	logs.Info("Logger start")

	err := godotenv.Load(".env")
	if err != nil {
		logs.Error("Fatal", "error loading .env file = ", err)
	}

	// инициализируем DB
	repo, err := db.NewDB(logs, cfg.AddrConDB)
	if err != nil {
		logs.Error("Fatal = not connect DB", "customerrors = ", err)
		panic(err)
	}
	logs.Info("DB connection")
	defer repo.Close()

	// инициализируем Service
	serv := service.NewService(repo, logs)
	logs.Info("Service run")

	// инициализируем проверку авторизацию
	serviceAuth := auth.NewServiceAuth([]byte(os.Getenv("TOKEN_SALT")), []byte(os.Getenv("PASSWORD_SALT")), repo)
	logs.Info("Service authorize run")

	// инициализируем запись Context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//инициализируем middleware
	authorization := middleware.NewAuthMiddleware(serviceAuth)

	// инициализируем Handlers
	registerHandler := register.NewHandlers(ctx, serviceAuth, logs)
	authorizeHandler := authorize.NewHandler(ctx, serviceAuth, logs)
	postOrderHandler := order.NewHandler(ctx, serv, logs)

	// инициализировали роутер и создали запросы
	r := chi.NewRouter()
	r.Use(middleware.WithLogging)

	r.Post("/api/user/register", registerHandler.ServerHTTP)
	r.Post("/api/user/login", authorizeHandler.ServerHTTP)
	r.Group(func(r chi.Router) {
		r.Use(authorization.ValidAuth)
		r.Post("/api/user/orders", postOrderHandler.Post)
		//r.GetUserByAccessToken("/api/user/order", handler.GetUserByAccessToken)
	})

	//r.Post("/api/user/withdrawals", handler.Post)
	//

	//r.GetUserByAccessToken("/api/user/balance", handler.GetUserByAccessToken)
	//r.GetUserByAccessToken("/api/user/balance/withdraw", handler.GetUserByAccessToken)

	if err := http.ListenAndServe(":8080", r); err != nil {
		logs.Error("Err:", logger.ErrAttr(err))
	}
}

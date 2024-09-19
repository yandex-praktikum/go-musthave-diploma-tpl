package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/handlers/authorize"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/handlers/balance"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/handlers/order"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/handlers/register"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/handlers/withdraw"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/middleware"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/auth"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/storage/db"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/workers"
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

	// запускаем воркер
	worker := workers.NewWorkerAccrual(serv, logs)

	//инициализируем middleware
	authorization := middleware.NewAuthMiddleware(serviceAuth)

	// инициализируем Handlers
	registerHandler := register.NewHandlers(ctx, serviceAuth, logs)
	authorizeHandler := authorize.NewHandler(ctx, serviceAuth, logs)
	ordersHandler := order.NewHandler(ctx, serv, logs)
	balanceHandler := balance.NewHandler(ctx, serv, logs)
	withdrawHandler := withdraw.NewHandler(ctx, serv, logs)

	// инициализировали роутер и создали запросы
	r := chi.NewRouter()
	r.Use(middleware.WithLogging)

	r.Post("/api/user/register", registerHandler.Post)
	r.Post("/api/user/login", authorizeHandler.Post)
	r.Group(func(r chi.Router) {
		r.Use(authorization.ValidAuth)
		r.Post("/api/user/orders", ordersHandler.Post)
		r.Get("/api/user/orders", ordersHandler.Get)
		r.Get("/api/user/balance", balanceHandler.Get)
		r.Post("/api/user/balance/withdraw", withdrawHandler.Post)
		r.Get("/api/user/withdrawals", withdrawHandler.Get)
	})

	go worker.StartWorkerAccrual(ctx, cfg.AccrualSystemAddress)
	if err := http.ListenAndServe(":8080", r); err != nil {
		logs.Error("Err:", logger.ErrAttr(err))
	}
}

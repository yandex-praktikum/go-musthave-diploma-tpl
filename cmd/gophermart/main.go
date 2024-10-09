package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"gophermart/internal/config"
	"gophermart/internal/handlers"
	"gophermart/internal/middleware"
	"gophermart/internal/repository"
	"gophermart/internal/service"
	"gophermart/storage"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	pgsStorage := &storage.PgStorage{}
	cfg, err := config.InitConfig()

	if err != nil {
		log.Fatalf("Error while initializing config: %v", err)
	}

	if cfg == nil {
		log.Fatalf("Error while initializing config: %v", err)
	}

	dsn := cfg.DatabaseDsn
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	err = db.AutoMigrate()

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	err = db.AutoMigrate(&storage.Order{})

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	err = db.AutoMigrate(&storage.User{})

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	err = db.AutoMigrate(&storage.UserBalance{})

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	err = db.AutoMigrate(&storage.Withdrawal{})

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	err = pgsStorage.Init(cfg.DatabaseDsn)

	if err != nil {
		log.Fatalf("Error while initializing db connection: %v", err)
	}

	defer pgsStorage.Close()

	userRepository := repository.UserRepository{
		DBStorage: pgsStorage,
	}
	userService := service.UserService{
		UserRepository: &userRepository,
	}
	orderRepository := repository.OrderRepository{
		DBStorage: pgsStorage,
	}
	orderService := service.OrderService{
		OrderRepository: &orderRepository,
	}
	withdrawRepository := repository.WithdrawRepository{
		DBStorage: pgsStorage,
	}
	withdrawService := service.WithdrawService{
		WithdrawRepository: &withdrawRepository,
	}
	userBalanceRepository := repository.UserBalanceRepository{
		DBStorage: pgsStorage,
	}
	userBalanceService := service.UserBalanceService{
		UserBalanceRepository: &userBalanceRepository,
	}
	userHandler := handlers.UserHandler{
		UserService:          userService,
		OrderService:         orderService,
		WithdrawService:      withdrawService,
		UserBalanceService:   userBalanceService,
		AccrualSystemAddress: cfg.AccrualSystemAddress,
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestDecompressor)
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", userHandler.Register)
		r.Post("/login", userHandler.Login)

		r.With(middleware.TokenAuthMiddleware).Route("/", func(r chi.Router) {
			r.Post("/orders", userHandler.SaveOrder)
			r.Get("/orders", userHandler.GetOrders)
			r.Get("/balance", userHandler.GetBalance)
			r.Post("/balance/withdraw", userHandler.Withdraw)
			r.Get("/withdrawals", userHandler.Withdrawals)
		})
	})

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %s", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Error shutting down server: %s", err)
	}

	log.Println("Server has exited.")
}

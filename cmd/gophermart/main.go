package main

import (
	"context"
	"log"
	"net/http"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/config"
	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/db"
	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/routers"
	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/service"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Не удалось инициализировать zap logger: %v", err)
	}
	defer func() {
		logger.Sync()
		if r := recover(); r != nil {
			logger.Fatal("Неожиданное завершение приложения", zap.Any("panic", r))
		}
	}()

	cfg := config.New()

	dbConn, err := db.Init(cfg.DatabaseURI)
	if err != nil {
		logger.Fatal("Ошибка подключения к БД", zap.Error(err))
	}
	defer func() {
		logger.Info("Закрытие соединения с БД")
		dbConn.Close()
	}()

	if err := db.Migrate(dbConn); err != nil {
		logger.Fatal("Ошибка миграции БД", zap.Error(err))
	}

	userRepo := db.NewUserRepoPG(dbConn)
	userService := service.NewUserService(userRepo)
	orderRepo := db.NewOrderRepoPG(dbConn)
	orderService := service.NewOrderService(orderRepo, userRepo)
	h := routers.NewHandler(userService, orderService, logger)
	r := routers.SetupRoutersWithLogger(h, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	orderService.StartOrderStatusWorker(ctx, cfg.AccrualSystemAddress, logger)

	logger.Info("Сервер запущен", zap.String("address", cfg.RunAddress))
	if err := http.ListenAndServe(cfg.RunAddress, r); err != nil {
		logger.Fatal("Ошибка запуска сервера", zap.Error(err))
	}
}

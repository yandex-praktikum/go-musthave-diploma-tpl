package main

import (
	"context"
	"fmt"
	cfg "musthave/internal/config/app"
	"musthave/internal/handler"
	srv "musthave/internal/service"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
)

func main() {
	lg := logrus.New()
	ctx := context.Background()
	cfg := cfg.NewConfig()              // инициализация конфига
	sh, err := srv.Create(ctx, lg, cfg) // инициализация сервиса
	if err != nil {
		panic(err)
	}

	h := handler.NewHandlers(sh)
	go h.StartHTTP(ctx, cfg.Port, cfg.SecretKey) // запуск сервера
	lg.Info("Listner: ", fmt.Sprintf("Запущен http слушатель на порту %s", cfg.Port))

	go func() {
		<-ctx.Done()
		h.StopHTTP(ctx)

	}()
	select {}
}

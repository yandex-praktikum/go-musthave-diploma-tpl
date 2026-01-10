package main

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
	cfg "github.com/tartushkin/go-musthave-diploma-tpl.git/internal/config/app"
	"github.com/tartushkin/go-musthave-diploma-tpl.git/internal/handler"
	srv "github.com/tartushkin/go-musthave-diploma-tpl.git/internal/service"
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

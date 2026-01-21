package main

import (
	"context"
	"log/slog"
	"musthave/internal/accrual"
	cfg "musthave/internal/config/app"
	"musthave/internal/handler"
	srv "musthave/internal/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// todo:
// - узнать как корректно обращаться к блэкбокс
// - дописать процесс получения статуса
// - безопасность мап
// - мапить стаутсы с блэкбокса на наши
// - тесты ?
func main() {
	lg := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	//ctx := context.Background()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := cfg.NewConfig() // инициализация конфига
	cl := accrual.NewClient(cfg.AccrualPath, cfg.ParamTimeOut)
	sh, err := srv.Create(ctx, lg, cfg, cl) // инициализация сервиса
	if err != nil {
		panic(err)
	}

	h := handler.NewHandlers(sh)

	go func() {
		if err := h.StartHTTP(ctx, cfg.Port, cfg.SecretKey); err != nil && err != http.ErrServerClosed {
			lg.Error("ошибка HTTP-сервера", "error", err)
			stop() // ускоряем остановку
		}
	}()

	lg.Info("HTTP-сервер запущен", "port", cfg.Port)

	// Ждём сигнала остановки
	<-ctx.Done()
	lg.Info("получен сигнал остановки, завершаем работу...")

}

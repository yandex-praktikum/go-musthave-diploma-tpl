package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/app/handler"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/app/middleware"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/service/auth"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/service/balance"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/service/orders"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/service/repository"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/worker"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/go-chi/chi"
	"github.com/s-lyash/go-musthave-diploma-tpl/config"
)

type Application struct {
	router *chi.Mux
	config *config.Config
	worker worker.WorkerIface
}

func InitApp(ctx context.Context) *Application {
	conf := config.ParseFlags()
	//connectionData := fmt.Sprintf("host=%s port=%d "+
	//	"dbname=%s ",
	//	"localhost", 5432, "postgres")

	db, err := sql.Open("pgx", conf.DB)
	if err != nil {
		fmt.Print(err)
	}

	repo := repository.NewRepository(db)
	authService := auth.NewService(repo.Auth, conf.SecretKey)
	balanceService := balance.NewService(repo.Balance)
	orderService := orders.NewService(repo.Orders, conf)

	middlewares := middleware.New(authService, conf)
	handle := handler.NewHandler(authService, balanceService, orderService, middlewares)
	router := handle.RoutesLink()
	work := worker.New(orderService)
	work.Start(ctx)

	return &Application{
		config: conf,
		router: router,
		worker: work,
	}
}

func (a *Application) StartApp(ctx context.Context, signal chan os.Signal) {
	srv := &http.Server{
		Addr:         a.config.ServerAddress,
		Handler:      a.router,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Print("server started at ", a.config.ServerAddress)
		}
	}()

	<-signal
	defer a.worker.Stop()
	if err := srv.Shutdown(ctx); err != nil {
	}
}

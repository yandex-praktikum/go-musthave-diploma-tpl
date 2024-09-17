package main

import (
	"flag"
	"github.com/caarlos0/env/v11"
	"github.com/korol8484/gofermart/internal/app/api/user_auth"
	"github.com/korol8484/gofermart/internal/app/cli/migrate"
	"github.com/korol8484/gofermart/internal/app/config"
	"github.com/korol8484/gofermart/internal/app/db"
	"github.com/korol8484/gofermart/internal/app/logger"
	"github.com/korol8484/gofermart/internal/app/server"
	"github.com/korol8484/gofermart/internal/app/token"
	"github.com/korol8484/gofermart/internal/app/user/auth"
	"github.com/korol8484/gofermart/internal/app/user/repository"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	os.Exit(initApp())
}

func initApp() int {
	cfg := &config.App{}

	flag.StringVar(&cfg.Listen, "a", ":8080", "Http service list addr")
	flag.StringVar(&cfg.DBDsn, "d", "", "set postgresql connection string (DSN)")
	flag.StringVar(&cfg.AccrualListen, "r", "", "Accrual system address")
	flag.Parse()

	zLog, err := logger.NewLogger(false)
	if err != nil {
		log.Printf("can't initalize logger %s", err)
		return 1
	}

	defer func(zLog *zap.Logger) {
		_ = zLog.Sync()
	}(zLog)

	if err = env.Parse(cfg); err != nil {
		zLog.Warn("can't parse environment variables", zap.Error(err))
	}

	if err = run(cfg, zLog); err != nil {
		zLog.Warn("can't run app", zap.Error(err))

		return 0
	}

	return 0
}

func run(cfg *config.App, log *zap.Logger) error {
	pg, err := db.NewPgDB(&db.Config{
		Dsn:             cfg.DBDsn,
		MaxIdleConn:     1,
		MaxOpenConn:     10,
		MaxLifetimeConn: time.Minute * 1,
	})
	if err != nil {
		return err
	}

	// Migrations
	mCmd, err := migrate.NewMigrateCmd(&migrate.Config{Dsn: cfg.DBDsn})
	if err != nil {
		return err
	}

	if err = mCmd.Up(); err != nil {
		return err
	}

	// Секрету тут не место, или через env или хранить как файл, тут упрощаем
	session := token.NewJwtService("secret", "Authorization", time.Hour*2)
	userRepo := repository.NewDbStore(pg)
	authSvc := auth.NewService(userRepo)
	authHandler := user_auth.NewAuthHandler(authSvc, session)

	httpServer := server.NewApp(&server.Config{
		Listen:         cfg.Listen,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}, log)

	httpServer.AddHandler(authHandler.RegisterRoutes())

	errCh := make(chan error, 1)
	oss, stop := make(chan os.Signal, 1), make(chan struct{}, 1)
	signal.Notify(oss, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-oss

		stop <- struct{}{}
	}()

	go func() {
		if err = httpServer.Run(false); err != nil {
			errCh <- err
		}
	}()

	for {
		select {
		case e := <-errCh:
			return e
		case <-stop:
			httpServer.Stop()
			return nil
		}
	}
}

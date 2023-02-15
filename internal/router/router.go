package router

import (
	"flag"

	"github.com/caarlos0/env"
	"github.com/labstack/echo"

	"GopherMart/internal/events"
)

type Config struct {
	ServerAddress  string `env:"RUN_ADDRESS"`
	BDAddress      string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

type serverMart struct {
	cfg  Config
	serv *echo.Echo
	db   events.DBI
}

func InitServer() *serverMart {
	return &serverMart{}
}

func (s serverMart) Router() error {
	if err := s.parseFlagCfg(); err != nil {
		return err
	}
	if err := s.connectDB(); err != nil {
		return err
	}

	e := echo.New()

	go s.updateAccrual()

	e.Use(s.gzip)

	e.POST("/api/user/registration", s.postAPIUserRegistration)
	e.POST("/api/user/login", s.postAPIUserLogin)

	e.Use(s.mwUserAuthentication)

	e.GET("/api/user/orders", s.getAPIUserOrders)           // Получение списка загруженных заказов
	e.GET("/api/user/balance", s.getAPIUserBalance)         // Получение текущего баланса пользователя
	e.GET("/api/user/withdrawals", s.getAPIUserWithdrawals) // Получение информации о выводе средств

	e.POST("/api/user/orders", s.postAPIUserOrders)                    // Загрузка номера заказа
	e.POST("/api/user/balance/withdraw", s.postAPIUserBalanceWithdraw) // Запрос на списание средств

	return nil
}

func (s *serverMart) parseFlagCfg() error {
	errConfig := env.Parse(&s.cfg)
	if errConfig != nil {
		return errConfig
	}
	if s.cfg.ServerAddress == "" {
		flag.StringVar(&s.cfg.ServerAddress, "a", "http://localhost:8080", "New RUN_ADDRESS")
	}
	if s.cfg.BDAddress == "" {
		flag.StringVar(&s.cfg.BDAddress, "d", "http://localhost:5432", "New DATABASE_URI")
	}
	if s.cfg.AccrualAddress == "" {
		flag.StringVar(&s.cfg.AccrualAddress, "r", "http://localhost:5433", "New ACCRUAL_SYSTEM_ADDRESS")
	}

	flag.Parse()
	return nil
}

func (s serverMart) connectDB() error {
	var err error
	if s.db, err = events.InitDB(); err != nil {
		return err
	}
	if err = s.db.Connect(s.cfg.BDAddress); err != nil {
		return err
	}
	return nil
}

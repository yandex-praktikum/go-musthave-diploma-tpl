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
	Cfg  Config
	Serv *echo.Echo
	DB   events.DBI
}

func InitServer() *serverMart {
	return &serverMart{}
}

func (s *serverMart) Router() error {
	if err := s.parseFlagCfg(); err != nil {
		return err
	}
	if err := s.connectDB(); err != nil {
		return err
	}

	e := echo.New()

	go s.updateAccrual()

	e.Use(s.gzip)

	e.POST("/api/user/register", s.postAPIUserRegister)
	e.POST("/api/user/login", s.postAPIUserLogin)

	//e.Use(s.mwUserAuthentication)

	e.GET("/api/user/orders", s.getAPIUserOrders, s.mwUserAuthentication)           // Получение списка загруженных заказов
	e.GET("/api/user/balance", s.getAPIUserBalance, s.mwUserAuthentication)         // Получение текущего баланса пользователя
	e.GET("/api/user/withdrawals", s.getAPIUserWithdrawals, s.mwUserAuthentication) // Получение информации о выводе средств

	e.POST("/api/user/orders", s.postAPIUserOrders, s.mwUserAuthentication)                    // Загрузка номера заказа
	e.POST("/api/user/balance/withdraw", s.postAPIUserBalanceWithdraw, s.mwUserAuthentication) // Запрос на списание средств

	errStart := e.Start(s.Cfg.ServerAddress)

	if errStart != nil {
		return errStart
	}
	return nil
}

func (s *serverMart) parseFlagCfg() error {
	errConfig := env.Parse(&s.Cfg)
	if errConfig != nil {
		return errConfig
	}
	if s.Cfg.ServerAddress == "" {
		flag.StringVar(&s.Cfg.ServerAddress, "a", "localhost:8080", "New RUN_ADDRESS")
	}
	if s.Cfg.BDAddress == "" {
		flag.StringVar(&s.Cfg.BDAddress, "d", "postgres://postgres:0000@localhost:5432/postgres", "New DATABASE_URI")
	}
	if s.Cfg.AccrualAddress == "" {
		flag.StringVar(&s.Cfg.AccrualAddress, "r", "", "New ACCRUAL_SYSTEM_ADDRESS")
	}

	flag.Parse()
	return nil
}

func (s *serverMart) connectDB() error {
	var err error
	if s.DB, err = events.InitDB(); err != nil {
		return err
	}
	if err = s.DB.Connect(s.Cfg.BDAddress); err != nil {
		return err
	}

	return nil
}

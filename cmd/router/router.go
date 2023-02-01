package router

import (
	"flag"

	"github.com/caarlos0/env"
	"github.com/labstack/echo"
)

type ConfigURL struct {
	ServerAddress  string `env:"RUN_ADDRESS"`
	BDAddress      string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

type serverMart struct {
	cfg  ConfigURL
	serv *echo.Echo
	db   DBI
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

	e.Use(s.gzip)
	//
	//e.POST("/api/user/register", s.postAPIUserRegister)
	//e.POST("/api/user/login", s.postAPIUserLogin)
	//
	//e.Use(s.CheakCookies)
	//
	//e.GET("/api/user/orders", s.getAPIUserOrders)
	//e.GET("/api/user/balance", s.getAPIUserBalance)
	//e.GET("/api/user/withdrawals", s.getAPIUserWithdrawals)
	//
	//e.POST("/api/user/orders", s.postAPIUserOrders)
	//e.POST("/api/user/balance/withdraw", s.postAPIUserBalanceWithdraw)

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
	if s.db, err = InitDB(); err != nil {
		return err
	}
	if err = s.db.Connect(s.cfg.BDAddress); err != nil {
		return err
	}
	return nil
}

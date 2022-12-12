package main

import (
	"github.com/gorilla/sessions"
	"github.com/iRootPro/gophermart/extservice/accrual"
	"github.com/iRootPro/gophermart/internal/apiserver"
	"github.com/iRootPro/gophermart/internal/store/sqlstore"
	"log"
)

const sessionKey = "SECRET_KEY"

func main() {
	config := apiserver.NewConfig()

	store := sqlstore.New()
	if err := store.Open(config.DatabaseURI); err != nil {
		log.Fatal(err)
	}

	if err := store.CreateTables(); err != nil {
		log.Fatal(err)
	}

	accrualConfig := accrual.NewAccrualConfig(config.AccrualSystemAddress, config.LogLevel)
	accrualService := accrual.NewAccrual(accrualConfig, store)

	go accrualService.Run()

	sessionsStore := sessions.NewCookieStore([]byte(sessionKey))
	s := apiserver.NewAPIServer(config, store, sessionsStore)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"github.com/gorilla/sessions"
	"github.com/iRootPro/gophermart/internal/apiserver"
	"log"
)

const sessionKey = "SECRET_KEY"

func main() {
	config := apiserver.NewConfig()
	sessionsStore := sessions.NewCookieStore([]byte(sessionKey))
	s := apiserver.NewAPIServer(config, sessionsStore)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}

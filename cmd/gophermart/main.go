package main

import (
	"github.com/iRootPro/gophermart/internal/apiserver"
	"log"
)

func main() {
	config := apiserver.NewConfig()
	s := apiserver.NewAPIServer(config)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}

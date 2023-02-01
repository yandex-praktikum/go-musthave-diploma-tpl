package main

import (
	"log"

	"GopherMart/cmd/router"
)

func main() {
	rout := router.InitServer()
	err := rout.Router()
	if err != nil {
		log.Fatal("Router:", err)
	}
}

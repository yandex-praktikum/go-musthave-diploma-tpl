package main

import "log"

func main() {
	rout := InitServer()
	err := rout.router()
	if err != nil {
		log.Fatal("Router:", err)
	}
}

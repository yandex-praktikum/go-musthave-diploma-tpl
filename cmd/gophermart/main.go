package main

import "github.com/evgfitil/gophermart.git/internal/logger"

func main() {
	if err := Execute(); err != nil {
		logger.Sugar.Fatalf("error starting server: %v", err)
	}
}

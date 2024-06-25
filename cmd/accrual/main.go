package main

import (
	"context"
	"fmt"
	"github.com/ShukinDmitriy/gophermart/cmd/accrual/fork"
	"github.com/ShukinDmitriy/gophermart/internal/config"
	"github.com/joho/godotenv"
	"os"
)

// init is invoked before main()
func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
}

func main() {
	conf := config.NewConfig()
	parseFlags(conf)
	parseEnvs(conf)

	envs := append(os.Environ(),
		"RUN_ADDRESS=127.0.0.1:8082",
		"DATABASE_URI=postgres://postgres:postgres@postgres:5432/gophermart_postgres?sslmode=disable",
	)

	flagAccrualBinaryPath := "/srv/www/go-course/gophermart/cmd/accrual/accrual_linux_amd64"
	p := fork.NewBackgroundProcess(context.Background(), flagAccrualBinaryPath,
		fork.WithEnv(envs...),
	)

	ctx := context.Background()

	err := p.Start(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Started")
	<-ctx.Done()
	fmt.Println("Ended")
}

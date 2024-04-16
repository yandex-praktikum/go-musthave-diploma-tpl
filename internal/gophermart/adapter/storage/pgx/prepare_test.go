package pgx_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

var postgresContainer *postgres.PostgresContainer

func TestMain(m *testing.M) {
	ctx, cancelFN := context.WithCancel(context.Background())
	defer func() {
		cancelFN()
	}()

	setup(ctx)
	code := m.Run()
	shutdown(ctx)
	os.Exit(code)
}

func shutdown(ctx context.Context) {
	if postgresContainer != nil {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}
}

func setup(ctx context.Context) {
	dbName := "users"
	dbUser := "user"
	dbPassword := "password"

	var err error
	postgresContainer, err = postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:15.2-alpine"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)

	if err != nil {
		log.Fatalf(err.Error())
	}

}

func createLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	log := logger.Sugar()
	return log
}

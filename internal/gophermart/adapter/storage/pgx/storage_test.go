package pgx_test

import (
	"context"
	"testing"

	"github.com/StasMerzlyakov/gophermart/internal/config"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/storage/pgx"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	//"github.com/testcontainers/testcontainers-go"
	//	"github.com/testcontainers/testcontainers-go/modules/postgres"
	//"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

func TestAuthFunctions(t *testing.T) {
	ctx, cancelFN := context.WithCancel(context.Background())
	/*postgresContainer, err := createContainer(ctx)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	connString, err := postgresContainer.ConnectionString(ctx) */

	connString := "postgres://postgres:postgres@localhost:5432/gophermarket"

	logger := createLogger()
	storage := pgx.NewStorage(ctx, logger, &config.GophermartConfig{
		MaxConns:    5,
		DatabaseUri: connString,
	})

	err := storage.Ping(ctx)
	require.NoError(t, err)

	login := "user"
	passHash := "hash"
	salt := "salt"
	ldata, err := storage.GetUserData(ctx, login)
	require.NoError(t, err)
	require.Nil(t, ldata)

	ldata = &domain.LoginData{
		Login: login,
		Hash:  passHash,
		Salt:  salt,
	}

	err = storage.RegisterUser(ctx, ldata)
	require.NoError(t, err)

	err = storage.RegisterUser(ctx, ldata)
	require.ErrorIs(t, err, domain.ErrLoginIsBusy)

	ldata, err = storage.GetUserData(ctx, login)
	require.NoError(t, err)
	require.NotNil(t, ldata)

	assert.Equal(t, login, ldata.Login)
	assert.Equal(t, passHash, ldata.Hash)
	assert.Equal(t, salt, ldata.Salt)
	assert.True(t, ldata.UserID > 0)

	cancelFN()
}

/*
func createContainer(ctx context.Context) (*postgres.PostgresContainer, error) {
	dbName := "users"
	dbUser := "user"
	dbPassword := "password"

	return postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:15.2-alpine"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)

} */

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

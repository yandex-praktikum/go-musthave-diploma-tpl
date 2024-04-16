package pgx_test

import (
	"context"
	"testing"

	"github.com/StasMerzlyakov/gophermart/internal/config"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/storage/pgx"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthFunctions(t *testing.T) {
	ctx, cancelFN := context.WithCancel(context.Background())

	defer cancelFN()

	connString, err := postgresContainer.ConnectionString(ctx)

	require.NoError(t, err)

	logger := createLogger()
	storage := pgx.NewStorage(ctx, logger, &config.GophermartConfig{
		MaxConns:    5,
		DatabaseUri: connString,
	})

	err = storage.Ping(ctx)
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

	_, err = storage.RegisterUser(ctx, ldata)
	require.NoError(t, err)

	_, err = storage.RegisterUser(ctx, ldata)
	require.ErrorIs(t, err, domain.ErrLoginIsBusy)

	ldata, err = storage.GetUserData(ctx, login)
	require.NoError(t, err)
	require.NotNil(t, ldata)

	assert.Equal(t, login, ldata.Login)
	assert.Equal(t, passHash, ldata.Hash)
	assert.Equal(t, salt, ldata.Salt)
	assert.True(t, ldata.UserID > 0)
}

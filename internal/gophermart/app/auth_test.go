package app_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/config"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/app"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/app/mocks"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRegisterSuccess(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockAuthStorage(ctrl)

	salt := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf}
	var saltFN = func(b []byte) (n int, err error) {
		require.Equal(t, 16, len(b))
		copy(b, salt)
		return 16, nil
	}

	saltB64 := base64.URLEncoding.EncodeToString(salt)

	auth := app.NewAuthFN(&config.GophermartConfig{
		TokenExp:    time.Second * 10,
		TokenSecret: "secret",
	}, mockStorage, saltFN)

	login := "user"
	pass := "!eraasd*{1}"
	regData := &domain.RegistrationData{
		Login:    login,
		Password: pass,
	}

	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(pass))
	hex := h.Sum(nil)
	hexB64 := base64.URLEncoding.EncodeToString(hex)

	mockStorage.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, loginData *domain.LoginData) {
		require.Equal(t, login, loginData.Login)
		require.Equal(t, saltB64, loginData.Salt)
		require.Equal(t, hexB64, loginData.Hash)
	}).Times(1)

	err := auth.Register(ctx, regData)
	require.NoError(t, err)
}

func TestRegisterLoginBusy(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockAuthStorage(ctrl)

	auth := app.NewAuth(&config.GophermartConfig{
		TokenExp:    time.Second * 10,
		TokenSecret: "secret",
	}, mockStorage)

	login := "user"
	pass := "!eraasd*{1}"
	regData := &domain.RegistrationData{
		Login:    login,
		Password: pass,
	}

	mockStorage.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, loginData *domain.LoginData) error {
		return fmt.Errorf("err %w", domain.ErrDataIsBusy)
	}).Times(1)

	err := auth.Register(ctx, regData)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrDataIsBusy)
}

func TestRegisterAnyError(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockAuthStorage(ctrl)

	auth := app.NewAuth(&config.GophermartConfig{
		TokenExp:    time.Second * 10,
		TokenSecret: "secret",
	}, mockStorage)

	login := "user"
	pass := "!eraasd*{1}"
	regData := &domain.RegistrationData{
		Login:    login,
		Password: pass,
	}

	mockStorage.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, loginData *domain.LoginData) error {
		return errors.New("custom error")
	}).Times(1)

	err := auth.Register(ctx, regData)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrServerInternal)
}

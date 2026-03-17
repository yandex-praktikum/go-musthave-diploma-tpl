package user

import (
	"context"
	"testing"
	"time"

	"gmart/internal/adapters/pgc"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"

	"go.uber.org/mock/gomock"
)

func TestAuthRepo_SignIn_UserNotFound_Metrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mPg := pgc.NewMockPgInstance(ctrl)
	mMetrics := NewMockAuthMetrics(ctrl)
	repo := NewAuthRepo(mPg, time.Hour, mMetrics)

	ctx := context.Background()

	// Настраиваем поведение: PgPool вызывает колбэк, который возвращает ErrNoRows
	mPg.EXPECT().
		PgPool(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context, pgc.PgxPoolIface) error) error {
			return pgx.ErrNoRows // Имитируем отсутствие юзера
		})

	// Ожидаем вызов метрики ошибки
	mMetrics.EXPECT().IncLoginError("err_user_not_found").Times(1)

	_, _, err := repo.SignIn(ctx, "admin", "password", "ua", "ip")

	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestAuthRepo_SignInSleep_TimingProtection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mPg := pgc.NewMockPgInstance(ctrl)
	// Метрики не важны для этого теста, передаем nil (проверка nil-safety)

	repo := NewAuthRepo(mPg, time.Hour, nil)

	// Фиксируем время, чтобы тест не зависел от фаз луны
	staticTime := time.Now()
	repo.now = func() time.Time { return staticTime }

	// Ставим среднее время bcrypt = 200мс
	repo.avgBcrypt.Store((200 * time.Millisecond).Nanoseconds())

	ctx := context.Background()

	// Имитируем быстрый провал в БД (юзер не найден за 5мс)
	mPg.EXPECT().PgPool(gomock.Any(), gomock.Any()).Return(pgx.ErrNoRows)

	start := time.Now()
	// Запускаем в горутине, чтобы проверить прерывание или задержку
	done := make(chan struct{})
	go func() {
		_, _, _ = repo.SignInSleep(ctx, "ghost", "pass", "ua", "ip")
		close(done)
	}()

	select {
	case <-done:
		elapsed := time.Since(start)
		// Проверяем, что мы спали минимум 150мс (200мс минус джиттер/задержки)
		assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(150), "Must sleep to protect against timing attack")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("SignInSleep took too long, protection logic might be broken")
	}
}

func TestAuthRepo_SignUp_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mPg := pgc.NewMockPgInstance(ctrl)
	mMetrics := NewMockAuthMetrics(ctrl)
	repo := NewAuthRepo(mPg, time.Hour, mMetrics)

	ctx := context.Background()

	// Используем Any(), так как время хеширования всегда разное
	mMetrics.EXPECT().
		ObserveBcrypt(gomock.Any()).
		Times(1)

	// Ожидаем вызов PgPool
	mPg.EXPECT().
		PgPool(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context, pgc.PgxPoolIface) error) error {
			// Имитируем успешное выполнение колбэка
			return nil
		})

	uid, err := repo.SignUp(ctx, "new_user", "password")

	assert.NoError(t, err)
	assert.Equal(t, int64(0), uid)
}

func TestAuthRepo_SignInSleep_ContextCancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mPg := pgc.NewMockPgInstance(ctrl)
	repo := NewAuthRepo(mPg, time.Hour, nil)
	repo.avgBcrypt.Store((500 * time.Millisecond).Nanoseconds())

	mPg.EXPECT().PgPool(gomock.Any(), gomock.Any()).Return(pgx.ErrNoRows)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, _, err := repo.SignInSleep(ctx, "ghost", "pass", "ua", "ip")

	elapsed := time.Since(start)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, elapsed.Milliseconds(), int64(100), "Should stop sleeping if context is cancelled")
}

package user

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gmart/internal/adapters/pgc"
	"gmart/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	mTx := pgc.NewMockPgxTxIface(ctrl)
	mockRow := pgc.NewMockRow(ctrl) // Создаем мок строки

	repo := NewAuthRepo(mPg, time.Hour, mMetrics)
	ctx := context.Background()

	expectedUID := domain.UserID(123)
	login := domain.Login("new_user")
	password := domain.Password("password")

	// 1. Метрики
	mMetrics.EXPECT().ObserveBcrypt(gomock.Any()).Times(1)

	// 2. Транзакция
	mPg.EXPECT().
		Tx(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context, pgc.PgxTxIface) error) error {
			return fn(ctx, mTx)
		})

	// 3. Scan внутри QueryRow
	// Важно: Scan принимает слайс интерфейсов. Нам нужно достать первый элемент.
	mockRow.EXPECT().
		Scan(gomock.Any()). // Ожидаем один аргумент (&userID)
		DoAndReturn(func(dest ...interface{}) error {
			if len(dest) > 0 {
				if ptr, ok := dest[0].(*domain.UserID); ok {
					*ptr = expectedUID
					return nil
				}
			}
			return fmt.Errorf("scan: wrong destination type")
		})

	// 4. QueryRow возвращает наш настроенный mockRow
	mTx.EXPECT().
		QueryRow(ctx, sqlInsertIntoUsers, login, gomock.Any()).
		Return(mockRow)

	// 5. Создание сессии
	mTx.EXPECT().
		Exec(ctx, sqlInsertIntoSessions, gomock.Any(), expectedUID, domain.UserAgent("UserAgent"), "IP", gomock.Any()).
		Return(pgconn.CommandTag{}, nil)

	// Вызов
	uid, sessionID, err := repo.SignUp(ctx, login, password, "UserAgent", "IP")

	// Проверки
	assert.NoError(t, err)
	assert.Equal(t, expectedUID, uid)
	assert.NotEqual(t, uuid.Nil, sessionID.UUID())
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

package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/b602op/go-musthave-diploma-tpl/internal/middleware"
	"github.com/b602op/go-musthave-diploma-tpl/internal/repository"
	"github.com/b602op/go-musthave-diploma-tpl/internal/service"
)

// fixture — тестовое окружение для HTTP-обработчиков: sqlmock БД,
// реальный стек сервисов и middleware аутентификации.
type fixture struct {
	mock sqlmock.Sqlmock
	auth *AuthHandler
	ord  *OrderHandler
	bal  *BalanceHandler
	mw   *middleware.AuthMiddleware
}

// newFixture создаёт обработчики на sqlmock без внешнего accrual-сервера.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	return newFixtureWithAccrual(t, "http://127.0.0.1:1")
}

// newFixtureWithAccrual создаёт обработчики с accrual-клиентом на указанном URL.
func newFixtureWithAccrual(t *testing.T, accrualURL string) *fixture {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	balanceRepo := repository.NewBalanceRepository(db)
	withdrawalRepo := repository.NewWithdrawalRepository(db)

	authService := service.NewAuthService(userRepo)
	orderService := service.NewOrderService(orderRepo, balanceRepo, service.NewAccrualClient(accrualURL))
	balanceService := service.NewBalanceService(balanceRepo, withdrawalRepo, db)

	mw := middleware.NewAuthMiddleware("test-secret")

	return &fixture{
		mock: mock,
		auth: NewAuthHandler(authService, mw),
		ord:  NewOrderHandler(orderService),
		bal:  NewBalanceHandler(balanceService),
		mw:   mw,
	}
}

// authedRequest создаёт запрос с валидным токеном пользователя userID.
func (f *fixture) authedRequest(t *testing.T, method, target string, userID int64) *http.Request {
	t.Helper()
	token, err := f.mw.GenerateToken(userID)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	req := httptest.NewRequest(method, target, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

// expectExists возвращает ожидание запроса проверки существования пользователя.
func (f *fixture) expectExists(login string, exists bool) {
	f.mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(login).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(exists))
}

// expectInsertUser возвращает ожидание вставки пользователя.
func (f *fixture) expectInsertUser(id int64) {
	f.mock.ExpectQuery(`INSERT INTO users`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
}

// expectFindUser возвращает ожидание поиска пользователя по логину.
func (f *fixture) expectFindUser(id int64, login, passwordHash string) {
	f.mock.ExpectQuery(`SELECT id, login, password, created_at FROM users`).
		WithArgs(login).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "created_at"}).
			AddRow(id, login, passwordHash, testTime()))
}

// serve выполняет запрос через middleware аутентификации и возвращает recorder.
func (f *fixture) serve(t *testing.T, h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	f.mw.Auth(h).ServeHTTP(rec, req)
	return rec
}

// testTime возвращает фиксированное время для тестов.
func testTime() time.Time {
	return time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
}

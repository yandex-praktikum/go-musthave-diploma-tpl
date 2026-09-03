package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/true18/go-musthave-diploma-tpl/internal/auth"
	"github.com/true18/go-musthave-diploma-tpl/internal/handler"
	"github.com/true18/go-musthave-diploma-tpl/internal/money"
	"github.com/true18/go-musthave-diploma-tpl/internal/repository"
	"github.com/true18/go-musthave-diploma-tpl/internal/server"
	"github.com/true18/go-musthave-diploma-tpl/internal/service"
)

const (
	validOrderA = "12345678903"
	validOrderB = "79927398713"
	invalidLuhn = "12345678904"
)

func TestRegisterSuccess(t *testing.T) {
	app, repo := newTestApp()

	response := register(t, app, "alice", "secret")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if findCookie(response) == nil {
		t.Fatal("auth cookie was not set")
	}
	user := repo.user(t, "alice")
	if bytes.Equal(user.PasswordHash, []byte("secret")) {
		t.Fatal("password was stored without hashing")
	}
}

func TestRegisterDuplicateLogin(t *testing.T) {
	app, _ := newTestApp()

	register(t, app, "alice", "secret")
	response := register(t, app, "alice", "another")

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestLoginSuccess(t *testing.T) {
	app, _ := newTestApp()
	register(t, app, "alice", "secret")

	response := doJSON(app, http.MethodPost, "/api/user/login", `{"login":"alice","password":"secret"}`, nil)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if findCookie(response) == nil {
		t.Fatal("auth cookie was not set")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	app, _ := newTestApp()
	register(t, app, "alice", "secret")

	response := doJSON(app, http.MethodPost, "/api/user/login", `{"login":"alice","password":"bad"}`, nil)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestPrivateRouteWithoutAuth(t *testing.T) {
	app, _ := newTestApp()

	response := doRequest(app, http.MethodGet, "/api/user/balance", nil, "", nil)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestUploadNewOrder(t *testing.T) {
	app, _ := newTestApp()
	cookie := registerCookie(t, app, "alice")

	response := doText(app, http.MethodPost, "/api/user/orders", validOrderA, cookie)

	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestUploadSameOrderBySameUser(t *testing.T) {
	app, _ := newTestApp()
	cookie := registerCookie(t, app, "alice")

	doText(app, http.MethodPost, "/api/user/orders", validOrderA, cookie)
	response := doText(app, http.MethodPost, "/api/user/orders", validOrderA, cookie)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestUploadSameOrderByAnotherUser(t *testing.T) {
	app, _ := newTestApp()
	alice := registerCookie(t, app, "alice")
	bob := registerCookie(t, app, "bob")

	doText(app, http.MethodPost, "/api/user/orders", validOrderA, alice)
	response := doText(app, http.MethodPost, "/api/user/orders", validOrderA, bob)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestUploadInvalidBody(t *testing.T) {
	app, _ := newTestApp()
	cookie := registerCookie(t, app, "alice")

	response := doText(app, http.MethodPost, "/api/user/orders", "abc", cookie)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestUploadInvalidLuhn(t *testing.T) {
	app, _ := newTestApp()
	cookie := registerCookie(t, app, "alice")

	response := doText(app, http.MethodPost, "/api/user/orders", invalidLuhn, cookie)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestUploadUnauthorized(t *testing.T) {
	app, _ := newTestApp()

	response := doText(app, http.MethodPost, "/api/user/orders", validOrderA, nil)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestListOrdersSortedNewestFirst(t *testing.T) {
	app, repo := newTestApp()
	cookie := registerCookie(t, app, "alice")
	userID := repo.user(t, "alice").ID

	oldTime := time.Date(2020, 12, 10, 12, 0, 0, 0, time.UTC)
	newTime := oldTime.Add(time.Hour)
	accrual := mustAmount(t, "500")
	repo.seedOrder(userID, "111", repository.StatusProcessing, nil, oldTime)
	repo.seedOrder(userID, "222", repository.StatusProcessed, &accrual, newTime)

	response := doRequest(app, http.MethodGet, "/api/user/orders", nil, "", cookie)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	var orders []map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &orders); err != nil {
		t.Fatal(err)
	}
	if len(orders) != 2 || orders[0]["number"] != "222" || orders[1]["number"] != "111" {
		t.Fatalf("orders are not sorted newest first: %v", orders)
	}
	if _, ok := orders[1]["accrual"]; ok {
		t.Fatalf("accrual must be omitted when nil: %v", orders[1])
	}
}

func TestListOrdersEmpty(t *testing.T) {
	app, _ := newTestApp()
	cookie := registerCookie(t, app, "alice")

	response := doRequest(app, http.MethodGet, "/api/user/orders", nil, "", cookie)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestEmptyBalance(t *testing.T) {
	app, _ := newTestApp()
	cookie := registerCookie(t, app, "alice")

	response := doRequest(app, http.MethodGet, "/api/user/balance", nil, "", cookie)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	assertJSONNumber(t, response.Body.Bytes(), "current", 0)
	assertJSONNumber(t, response.Body.Bytes(), "withdrawn", 0)
}

func TestBalanceAfterProcessedAccrual(t *testing.T) {
	app, repo := newTestApp()
	cookie := registerCookie(t, app, "alice")
	userID := repo.user(t, "alice").ID
	amount := mustAmount(t, "500.5")
	repo.seedOrder(userID, "111", repository.StatusProcessed, &amount, time.Now())

	response := doRequest(app, http.MethodGet, "/api/user/balance", nil, "", cookie)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	assertJSONNumber(t, response.Body.Bytes(), "current", 500.5)
	assertJSONNumber(t, response.Body.Bytes(), "withdrawn", 0)
}

func TestWithdrawSuccess(t *testing.T) {
	app, repo := newTestApp()
	cookie := registerCookie(t, app, "alice")
	userID := repo.user(t, "alice").ID
	amount := mustAmount(t, "1000")
	repo.seedOrder(userID, "111", repository.StatusProcessed, &amount, time.Now())

	response := doJSON(app, http.MethodPost, "/api/user/balance/withdraw", `{"order":"`+validOrderB+`","sum":751}`, cookie)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}

	response = doRequest(app, http.MethodGet, "/api/user/balance", nil, "", cookie)
	assertJSONNumber(t, response.Body.Bytes(), "current", 249)
	assertJSONNumber(t, response.Body.Bytes(), "withdrawn", 751)
}

func TestWithdrawInsufficientFunds(t *testing.T) {
	app, _ := newTestApp()
	cookie := registerCookie(t, app, "alice")

	response := doJSON(app, http.MethodPost, "/api/user/balance/withdraw", `{"order":"`+validOrderB+`","sum":751}`, cookie)

	if response.Code != http.StatusPaymentRequired {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestWithdrawInvalidOrder(t *testing.T) {
	app, _ := newTestApp()
	cookie := registerCookie(t, app, "alice")

	response := doJSON(app, http.MethodPost, "/api/user/balance/withdraw", `{"order":"`+invalidLuhn+`","sum":1}`, cookie)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestWithdrawalsEmpty(t *testing.T) {
	app, _ := newTestApp()
	cookie := registerCookie(t, app, "alice")

	response := doRequest(app, http.MethodGet, "/api/user/withdrawals", nil, "", cookie)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestWithdrawalsSortedNewestFirst(t *testing.T) {
	app, repo := newTestApp()
	cookie := registerCookie(t, app, "alice")
	userID := repo.user(t, "alice").ID
	amount := mustAmount(t, "1000")
	repo.seedOrder(userID, "111", repository.StatusProcessed, &amount, time.Now())

	firstTime := time.Date(2020, 12, 9, 10, 0, 0, 0, time.UTC)
	secondTime := firstTime.Add(time.Hour)
	repo.now = func() time.Time { return firstTime }
	doJSON(app, http.MethodPost, "/api/user/balance/withdraw", `{"order":"`+validOrderA+`","sum":10}`, cookie)
	repo.now = func() time.Time { return secondTime }
	doJSON(app, http.MethodPost, "/api/user/balance/withdraw", `{"order":"`+validOrderB+`","sum":20}`, cookie)

	response := doRequest(app, http.MethodGet, "/api/user/withdrawals", nil, "", cookie)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	var withdrawals []map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &withdrawals); err != nil {
		t.Fatal(err)
	}
	if len(withdrawals) != 2 || withdrawals[0]["order"] != validOrderB || withdrawals[1]["order"] != validOrderA {
		t.Fatalf("withdrawals are not sorted newest first: %v", withdrawals)
	}
}

func newTestApp() (http.Handler, *memoryRepo) {
	repo := newMemoryRepo()
	appService := service.New(repo)
	authManager := auth.New([]byte("test-secret"))
	appHandler := handler.New(appService, authManager)
	appServer := server.New(":0", appHandler, authManager, zap.NewNop())
	return appServer.Handler, repo
}

func register(t *testing.T, app http.Handler, login string, password string) *httptest.ResponseRecorder {
	t.Helper()
	return doJSON(app, http.MethodPost, "/api/user/register", `{"login":"`+login+`","password":"`+password+`"}`, nil)
}

func registerCookie(t *testing.T, app http.Handler, login string) *http.Cookie {
	t.Helper()
	response := register(t, app, login, "secret")
	if response.Code != http.StatusOK {
		t.Fatalf("register status = %d", response.Code)
	}
	cookie := findCookie(response)
	if cookie == nil {
		t.Fatal("auth cookie was not set")
	}
	return cookie
}

func findCookie(response *httptest.ResponseRecorder) *http.Cookie {
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == auth.CookieName {
			return cookie
		}
	}
	return nil
}

func doJSON(app http.Handler, method string, path string, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	return doRequest(app, method, path, bytes.NewBufferString(body), "application/json", cookie)
}

func doText(app http.Handler, method string, path string, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	return doRequest(app, method, path, bytes.NewBufferString(body), "text/plain", cookie)
}

func doRequest(app http.Handler, method string, path string, body *bytes.Buffer, contentType string, cookie *http.Cookie) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader(body.Bytes())
	}
	request := httptest.NewRequest(method, path, reader)
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	return response
}

func assertJSONNumber(t *testing.T, data []byte, field string, want float64) {
	t.Helper()

	var response map[string]float64
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	if response[field] != want {
		t.Fatalf("%s = %v, want %v in %s", field, response[field], want, data)
	}
}

func mustAmount(t *testing.T, value string) money.Amount {
	t.Helper()

	amount, err := money.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	return amount
}

type memoryRepo struct {
	mu          sync.Mutex
	nextUserID  int64
	users       map[string]repository.User
	orders      map[string]repository.Order
	withdrawals map[int64][]repository.Withdrawal
	now         func() time.Time
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		nextUserID:  1,
		users:       make(map[string]repository.User),
		orders:      make(map[string]repository.Order),
		withdrawals: make(map[int64][]repository.Withdrawal),
		now:         time.Now,
	}
}

func (r *memoryRepo) CreateUser(_ context.Context, login string, passwordHash []byte) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[login]; ok {
		return 0, repository.ErrUserExists
	}
	user := repository.User{
		ID:           r.nextUserID,
		Login:        login,
		PasswordHash: append([]byte(nil), passwordHash...),
		CreatedAt:    r.now(),
	}
	r.nextUserID++
	r.users[login] = user
	return user.ID, nil
}

func (r *memoryRepo) FindUserByLogin(_ context.Context, login string) (repository.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[login]
	if !ok {
		return repository.User{}, repository.ErrNotFound
	}
	return user, nil
}

func (r *memoryRepo) CreateOrder(_ context.Context, userID int64, number string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[number]
	if ok && order.UserID == userID {
		return repository.ErrOrderExistsSameUser
	}
	if ok {
		return repository.ErrOrderExistsAnotherUser
	}
	r.orders[number] = repository.Order{
		Number:     number,
		UserID:     userID,
		Status:     repository.StatusNew,
		UploadedAt: r.now(),
	}
	return nil
}

func (r *memoryRepo) ListUserOrders(_ context.Context, userID int64) ([]repository.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var orders []repository.Order
	for _, order := range r.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].UploadedAt.After(orders[j].UploadedAt)
	})
	return orders, nil
}

func (r *memoryRepo) ListPendingOrders(_ context.Context, limit int) ([]repository.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var orders []repository.Order
	for _, order := range r.orders {
		if order.Status == repository.StatusNew || order.Status == repository.StatusProcessing {
			orders = append(orders, order)
		}
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].UploadedAt.Before(orders[j].UploadedAt)
	})
	if len(orders) > limit {
		orders = orders[:limit]
	}
	return orders, nil
}

func (r *memoryRepo) UpdateOrder(_ context.Context, number string, status string, accrual *money.Amount) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[number]
	if !ok {
		return repository.ErrNotFound
	}
	if order.Status == repository.StatusInvalid || order.Status == repository.StatusProcessed {
		return nil
	}
	order.Status = status
	order.Accrual = accrual
	r.orders[number] = order
	return nil
}

func (r *memoryRepo) Balance(_ context.Context, userID int64) (repository.Balance, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.balance(userID), nil
}

func (r *memoryRepo) Withdraw(_ context.Context, userID int64, orderNumber string, amount money.Amount) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	balance := r.balance(userID)
	if balance.Current < amount {
		return repository.ErrInsufficientFunds
	}
	r.withdrawals[userID] = append(r.withdrawals[userID], repository.Withdrawal{
		OrderNumber: orderNumber,
		Sum:         amount,
		ProcessedAt: r.now(),
	})
	return nil
}

func (r *memoryRepo) ListWithdrawals(_ context.Context, userID int64) ([]repository.Withdrawal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	withdrawals := append([]repository.Withdrawal(nil), r.withdrawals[userID]...)
	sort.Slice(withdrawals, func(i, j int) bool {
		return withdrawals[i].ProcessedAt.After(withdrawals[j].ProcessedAt)
	})
	return withdrawals, nil
}

func (r *memoryRepo) seedOrder(userID int64, number string, status string, accrual *money.Amount, uploadedAt time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[number] = repository.Order{
		Number:     number,
		UserID:     userID,
		Status:     status,
		Accrual:    accrual,
		UploadedAt: uploadedAt,
	}
}

func (r *memoryRepo) user(t *testing.T, login string) repository.User {
	t.Helper()

	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[login]
	if !ok {
		t.Fatalf("user %q not found", login)
	}
	return user
}

func (r *memoryRepo) balance(userID int64) repository.Balance {
	var accrued money.Amount
	var withdrawn money.Amount
	for _, order := range r.orders {
		if order.UserID == userID && order.Status == repository.StatusProcessed && order.Accrual != nil {
			accrued += *order.Accrual
		}
	}
	for _, withdrawal := range r.withdrawals[userID] {
		withdrawn += withdrawal.Sum
	}
	return repository.Balance{Current: accrued - withdrawn, Withdrawn: withdrawn}
}

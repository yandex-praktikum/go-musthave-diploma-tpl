package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Raime-34/gophermart.git/internal/dto"
	"github.com/Raime-34/gophermart.git/internal/gophermart"
	mocksgophermart "github.com/Raime-34/gophermart.git/internal/server/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestServer_registerUser_BadJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	ch := mocksgophermart.NewMockcookieHandlerInterface(ctrl)
	s := &Server{gophermart: gm, cookieHandler: ch}

	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()

	s.registerUser(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServer_registerUser_Conflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	ch := mocksgophermart.NewMockcookieHandlerInterface(ctrl)
	s := &Server{gophermart: gm, cookieHandler: ch}

	cred := dto.UserCredential{Login: "u", Password: "p"}
	gm.EXPECT().
		RegisterUser(gomock.Any(), cred).
		Return(errors.New("dup"))

	body, _ := json.Marshal(cred)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.registerUser(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

func TestServer_registerUser_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	ch := mocksgophermart.NewMockcookieHandlerInterface(ctrl)
	s := &Server{gophermart: gm, cookieHandler: ch}

	cred := dto.UserCredential{Login: "u", Password: "p"}
	gm.EXPECT().
		RegisterUser(gomock.Any(), cred).
		Return(nil)

	body, _ := json.Marshal(cred)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.registerUser(w, req)

	require.Equal(t, http.StatusOK, w.Code) // не писали header => 200
}

func TestServer_loginUser_BadJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	ch := mocksgophermart.NewMockcookieHandlerInterface(ctrl)
	s := &Server{gophermart: gm, cookieHandler: ch}

	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()

	s.loginUser(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServer_loginUser_Unauthorized_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	ch := mocksgophermart.NewMockcookieHandlerInterface(ctrl)
	s := &Server{gophermart: gm, cookieHandler: ch}

	cred := dto.UserCredential{Login: "u", Password: "p"}
	gm.EXPECT().
		LoginUser(gomock.Any(), cred).
		Return(nil, gophermart.ErrUserNotFound)

	body, _ := json.Marshal(cred)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.loginUser(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Empty(t, w.Result().Cookies())
}

func TestServer_loginUser_Unauthorized_IncorrectPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	ch := mocksgophermart.NewMockcookieHandlerInterface(ctrl)
	s := &Server{gophermart: gm, cookieHandler: ch}

	cred := dto.UserCredential{Login: "u", Password: "p"}
	gm.EXPECT().
		LoginUser(gomock.Any(), cred).
		Return(nil, gophermart.ErrIncorrectPassword)

	body, _ := json.Marshal(cred)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.loginUser(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Empty(t, w.Result().Cookies())
}

func TestServer_loginUser_OK_SetsCookie_AndStoresSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	ch := mocksgophermart.NewMockcookieHandlerInterface(ctrl)
	s := &Server{gophermart: gm, cookieHandler: ch}

	cred := dto.UserCredential{Login: "u", Password: "p"}
	user := &dto.UserData{Login: "u"} // подстрой под твою структуру

	gm.EXPECT().
		LoginUser(gomock.Any(), cred).
		Return(user, nil)

	// sid генерится случайно, поэтому матчим любой непустой
	ch.EXPECT().
		Set(gomock.Any(), user).
		Do(func(sid string, _ *dto.UserData) {
			require.NotEmpty(t, sid)
			require.Len(t, sid, 64) // 32 байта -> hex = 64 символа
		})

	body, _ := json.Marshal(cred)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.loginUser(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, "sid", cookies[0].Name)
	require.NotEmpty(t, cookies[0].Value)
	require.Len(t, cookies[0].Value, 64)
}

func TestServer_registerOrder_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	// ReadAll вернет ошибку, если Body=nil
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", nil)
	req.Body = nil

	w := httptest.NewRecorder()
	s.registerOrder(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServer_registerOrder_InvalidLuhn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	// 123 не проходит луна
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("123"))
	w := httptest.NewRecorder()

	s.registerOrder(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestServer_registerOrder_InsertError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	// валидный по Луну пример
	order := "79927398713"

	gm.EXPECT().
		InsertOrder(gomock.Any(), order).
		Return(errors.New("db down"))

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(order))
	w := httptest.NewRecorder()

	s.registerOrder(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestServer_registerOrder_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	order := "79927398713"
	gm.EXPECT().
		InsertOrder(gomock.Any(), order).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(order))
	w := httptest.NewRecorder()

	s.registerOrder(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestServer_getOrders_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	gm.EXPECT().
		GetUserOrders(gomock.Any()).
		Return(nil, errors.New("fail"))

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	w := httptest.NewRecorder()

	s.getOrders(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestServer_getOrders_NoContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	gm.EXPECT().
		GetUserOrders(gomock.Any()).
		Return([]*dto.GetOrdersInfoResp{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	w := httptest.NewRecorder()

	s.getOrders(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestServer_getOrders_OK_JSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	orders := []*dto.GetOrdersInfoResp{
		{Number: "1", Status: "NEW"},
	}
	gm.EXPECT().
		GetUserOrders(gomock.Any()).
		Return(orders, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	w := httptest.NewRecorder()

	s.getOrders(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var got []*dto.GetOrdersInfoResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Equal(t, orders, got)
}

func TestServer_registerWithdrawl_BadJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()

	s.registerWithdrawl(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServer_registerWithdrawl_InvalidLuhn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	body, _ := json.Marshal(dto.WithdrawRequest{Order: "123", Sum: 10})
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.registerWithdrawl(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestServer_registerWithdrawl_NotEnoughBonuses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	wreq := dto.WithdrawRequest{Order: "79927398713", Sum: 10} // валидный по Луну
	gm.EXPECT().
		ProcessWithdraw(gomock.Any(), wreq).
		Return(gophermart.ErrNotEnoughBonuses)

	body, _ := json.Marshal(wreq)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.registerWithdrawl(w, req)

	require.Equal(t, http.StatusPaymentRequired, w.Code)
}

func TestServer_registerWithdrawl_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	wreq := dto.WithdrawRequest{Order: "79927398713", Sum: 10}
	gm.EXPECT().
		ProcessWithdraw(gomock.Any(), wreq).
		Return(nil)

	body, _ := json.Marshal(wreq)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.registerWithdrawl(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestServer_getWithdrawls_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	gm.EXPECT().
		GetWithdraws(gomock.Any()).
		Return(nil, errors.New("fail"))

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	w := httptest.NewRecorder()

	s.getWithdrawls(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestServer_getWithdrawls_NoContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	gm.EXPECT().
		GetWithdraws(gomock.Any()).
		Return([]*dto.WithdrawInfo{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	w := httptest.NewRecorder()

	s.getWithdrawls(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestServer_getWithdrawls_OK_JSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	withs := []*dto.WithdrawInfo{
		{Order: "79927398713", Sum: 10},
	}
	gm.EXPECT().
		GetWithdraws(gomock.Any()).
		Return(withs, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	w := httptest.NewRecorder()

	s.getWithdrawls(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var got []*dto.WithdrawInfo
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Equal(t, withs, got)
}

func TestServer_getBalance_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	gm.EXPECT().
		GetUserBalance(gomock.Any()).
		Return(nil, errors.New("fail"))

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	w := httptest.NewRecorder()

	s.getBalance(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestServer_getBalance_OK_JSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gm := mocksgophermart.NewMockgophermartInterface(ctrl)
	s := &Server{gophermart: gm}

	bal := &dto.BalanceInfo{Current: 100, Withdraw: 20} // подстрой под поля
	gm.EXPECT().
		GetUserBalance(gomock.Any()).
		Return(bal, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	w := httptest.NewRecorder()

	s.getBalance(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var got dto.BalanceInfo
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Equal(t, *bal, got)
}

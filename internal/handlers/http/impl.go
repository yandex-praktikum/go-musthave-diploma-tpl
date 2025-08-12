package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/serviceErorrs"
)

type userClient interface {
	RegisterAndLoginUser(ctx context.Context, reg domain.Login) (domain.SessionToken, error)
	LoginUser(ctx context.Context, reg domain.Login) (domain.SessionToken, error)
}

type Implementation struct {
	a userClient
}

func New(a userClient) *Implementation {
	return &Implementation{
		a: a,
	}
}

type login struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register POST /api/user/register — регистрация пользователя;
func (i *Implementation) Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		appErr := serviceErorrs.AppErrorFromError(err).LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
		return
	}

	l := login{}
	if err = json.Unmarshal(body, &l); err != nil {
		appErr := serviceErorrs.NewBadRequest().LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
		return
	}

	authInfo, err := i.a.RegisterAndLoginUser(r.Context(), domain.Login{
		Login:    l.Login,
		Password: l.Password,
	})
	if err != nil {
		appErr := serviceErorrs.AppErrorFromError(err).LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    AuthCookiesName,
		Value:   authInfo.Token,
		Expires: time.Now().Add(24 * 180 * time.Hour),
		Secure:  true,
	})

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte("success register and authorize")); err != nil {
		appErr := serviceErorrs.AppErrorFromError(err).LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
	}

	return
}

// Login POST /api/user/login — аутентификация пользователя;
func (i *Implementation) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		appErr := serviceErorrs.AppErrorFromError(err).LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
		return
	}

	l := login{}
	if err = json.Unmarshal(body, &l); err != nil {
		appErr := serviceErorrs.NewBadRequest().LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
		return
	}

	authInfo, err := i.a.LoginUser(r.Context(), domain.Login{
		Login:    l.Login,
		Password: l.Password,
	})
	if err != nil {
		appErr := serviceErorrs.AppErrorFromError(err).LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    AuthCookiesName,
		Value:   authInfo.Token,
		Expires: time.Now().Add(24 * 180 * time.Hour),
		Secure:  true,
	})

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte("success authorize")); err != nil {
		appErr := serviceErorrs.AppErrorFromError(err).LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
	}

	return
}

// AddOrder POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
func (i *Implementation) AddOrder(w http.ResponseWriter, r *http.Request) {
	return
}

// GetOrders GET /api/user/orders — получение списка загруженных пользователем номеров заказов,
// статусов их обработки и информации о начислениях;
// TODO пагинация
func (i *Implementation) GetOrders(w http.ResponseWriter, r *http.Request) {
	return
}

// GetBalance GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
func (i *Implementation) GetBalance(w http.ResponseWriter, r *http.Request) {
	return
}

// WithdrawalPoints POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
func (i *Implementation) WithdrawalPoints(w http.ResponseWriter, r *http.Request) {
	return
}

// GetWithdrawals GET /api/user/withdrawals — получение информации о выводе средств с накопительного счёта пользователем.
// TODO пагинация
func (i *Implementation) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	return
}

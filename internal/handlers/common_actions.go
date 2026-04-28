package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/KalessinD/gophermart/internal/common"
	middleware "github.com/KalessinD/gophermart/internal/middleware"
	model "github.com/KalessinD/gophermart/internal/models"
	service "github.com/KalessinD/gophermart/internal/services"
)

const (
	TokenExpiration = time.Hour * 3
	CookieTokenName = "token"
	ParentPath      = "/api/user/"
)

type (
	CommonHandler struct {
		commonService service.CommonActionsInterface
		authService   service.AuthInterface
	}

	CommonHandlerInterface interface {
		Login(w http.ResponseWriter, r *http.Request)
		Register(w http.ResponseWriter, r *http.Request)
	}
)

/*
Конструктор для хендлеров работающих без автоиизации
*/
func NewCommonHandler(commonService service.CommonActionsInterface, authService service.AuthInterface) CommonHandlerInterface {
	return &CommonHandler{
		commonService: commonService,
		authService:   authService,
	}
}

/*
Аутентификация пользователя

Формат запроса:
```
POST /api/user/login HTTP/1.1
Content-Type: application/json
...

	{
	    "login": "<login>",
	    "password": "<password>"
	}

```

Возможные коды ответа:
- 200 — пользователь успешно аутентифицирован;
- 400 — неверный формат запроса;
- 401 — неверная пара логин/пароль;
- 500 — внутренняя ошибка сервера.
*/
func (h *CommonHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := middleware.GetLogger(ctx).Sugar()
	user, err := h.commonChecks(w, r)
	if err != nil {
		log.Debugf("login failed: %s", err.Error())
		return
	}

	if err = h.commonService.Login(ctx, user); err != nil {
		status := h.defineResponseStatusByError(err)
		if status == http.StatusInternalServerError {
			log.Errorf("login failed: %s", err.Error())
		} else {
			log.Debugf("login failed: %s", err.Error())
		}

		w.WriteHeader(status)

		return
	}

	if err := h.setAuthCookie(w, user); err != nil {
		log.Errorf("setting auth cookie failed: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/*
Регистрация пользователя

Формат запроса:
```
POST /api/user/register HTTP/1.1
Content-Type: application/json
...

	{
	    "login": "<login>",
	    "password": "<password>"
	}

```
Возможные коды ответа:
- 200 — пользователь успешно зарегистрирован и аутентифицирован;
- 400 — неверный формат запроса;
- 409 — логин уже занят;
- 500 — внутренняя ошибка сервера.
*/
func (h *CommonHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := middleware.GetLogger(ctx).Sugar()
	user, err := h.commonChecks(w, r)
	if err != nil {
		log.Debugf("user registration failed: %s", err.Error())
		return
	}

	if err = h.commonService.Register(ctx, user); err != nil {
		status := h.defineResponseStatusByError(err)
		if status == http.StatusInternalServerError {
			log.Errorf("user registration failed: %s", err.Error())
		} else {
			log.Debugf("user registration failed: %s", err.Error())
		}

		w.WriteHeader(status)

		return
	}

	if err := h.setAuthCookie(w, user); err != nil {
		log.Errorf("setting auth cookie failed: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *CommonHandler) commonChecks(w http.ResponseWriter, r *http.Request) (user *model.User, err error) {
	contentType := r.Header.Get("Content-Type")
	if contentType != common.AppJSONContentType {
		err = fmt.Errorf("bad request: %s", "wrong content-type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("bad request: %w", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err = model.FromJSON[model.User](body)
	if err != nil {
		err = fmt.Errorf("can't parse request body: %w", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	return
}

func (h *CommonHandler) defineResponseStatusByError(err error) (status int) {
	switch {
	case errors.Is(err, model.ErrWrongLoginLength):
		status = http.StatusBadRequest
	case errors.Is(err, model.ErrWrongPasswordLength):
		status = http.StatusBadRequest
	case errors.Is(err, model.ErrUserNotFound):
		status = http.StatusUnauthorized
	case errors.Is(err, model.ErrWrongPassword):
		status = http.StatusUnauthorized
	case errors.Is(err, model.ErrUserExists):
		status = http.StatusConflict
	default:
		status = http.StatusInternalServerError
	}

	return
}

func (h *CommonHandler) setAuthCookie(w http.ResponseWriter, user *model.User) error {
	expireAt := time.Now().Add(TokenExpiration)
	tokenString, err := h.authService.GenerateToken(user, expireAt)
	if err != nil {
		return fmt.Errorf("token generation failed: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CookieTokenName,
		Value:    tokenString,
		Expires:  expireAt,
		Secure:   false,                // false только ради разработки
		HttpOnly: true,                 // Защита от XSS (JS не может прочитать куку)
		Path:     ParentPath,           // Доступна в /api/user
		SameSite: http.SameSiteLaxMode, // Защита от CSRF
	})

	return nil
}

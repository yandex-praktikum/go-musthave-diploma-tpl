package auth

import (
	"encoding/json"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/models"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/service/auth"
)

type Handler struct {
	authService auth.Auth
}

func NewHandler(auth auth.Auth) *Handler {
	return &Handler{
		authService: auth,
	}
}

func (h Handler) registration(w http.ResponseWriter, r *http.Request) {
	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, models.ErrInvalidBody, http.StatusBadRequest)
		return
	}
	if len(user.Login) <= 0 || len(user.Password) <= 0 {
		http.Error(w, models.ErrInvalidBody, http.StatusBadRequest)
		return
	}
	if user.Login == "" || user.Password == "" {
		http.Error(w, models.ErrInvalidBody, http.StatusBadRequest)
		return
	}

	session, err := h.authService.Registration(user)
	if err != nil && err.Error() == models.EXIST {
		w.WriteHeader(LoginIsTaken)
	} else if err != nil {
		w.WriteHeader(InternalServerError)
	}

	cookie := &http.Cookie{
		Name:  models.COOKIE,
		Value: session,
	}

	http.SetCookie(w, cookie)
	w.Header().Set(models.CONTENT, models.APPJSON)
	w.WriteHeader(SuccessAuth)
}

func (h Handler) login(w http.ResponseWriter, r *http.Request) {
	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, models.ErrInvalidBody, http.StatusBadRequest)
		return
	}

	if len(user.Login) <= 0 || len(user.Password) <= 0 {
		http.Error(w, models.ErrInvalidBody, http.StatusBadRequest)
		return
	}

	if user.Login == "" || user.Password == "" {
		http.Error(w, models.ErrInvalidBody, http.StatusBadRequest)
		return
	}

	session, err := h.authService.Login(user)
	if err != nil {
		w.WriteHeader(WrongLoginPassword)
	}

	cookie := &http.Cookie{
		Name:  models.COOKIE,
		Value: session,
	}

	http.SetCookie(w, cookie)

	w.Header().Set(models.CONTENT, models.APPJSON)
	w.WriteHeader(SuccessAuth)
}

func (h Handler) AuthRouter(r chi.Router) {
	r.Post("/register", h.registration)
	r.Post("/login", h.login)
}

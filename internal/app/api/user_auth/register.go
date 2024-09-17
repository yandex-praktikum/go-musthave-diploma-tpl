package user_auth

import (
	"encoding/json"
	"errors"
	"github.com/korol8484/gofermart/internal/app/domain"
	"io"
	"net/http"
)

func (h *Handler) registerHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	req := &request{}
	if err = json.Unmarshal(body, req); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	// можно добавить сервис валидации, политики пароля...
	if req.Login == "" || req.Password == "" {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	u := &domain.User{
		Login: req.Login,
	}

	if u, err = h.auth.CreateUser(r.Context(), u, req.Password); err != nil {
		if errors.Is(err, domain.ErrIssetUser) {
			http.Error(w, "логин уже занят", http.StatusConflict)
			return
		}

		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if err = h.session.CreateSession(w, r, u.ID); err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

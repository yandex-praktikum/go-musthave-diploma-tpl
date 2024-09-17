package user_auth

import (
	"encoding/json"
	"io"
	"net/http"
)

func (h *Handler) authHandler(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.auth.Auth(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, "неверная пара логин/пароль", http.StatusUnauthorized)
		return
	}

	if err = h.session.CreateSession(w, r, user.ID); err != nil {
		http.Error(w, "внутренняя ошибка сервера.", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

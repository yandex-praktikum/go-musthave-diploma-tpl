package handlers

import (
	"encoding/json"
	"github.com/botaevg/gophermart/internal/config"
	"github.com/botaevg/gophermart/internal/models"
	"github.com/botaevg/gophermart/internal/service"
	"io"
	"net/http"
)

type handler struct {
	config config.Config
	auth   service.Auth
}

func NewHandler(config config.Config, auth service.Auth) *handler {
	return &handler{
		config: config,
		auth:   auth,
	}
}

func (h *handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var UserAPI models.UserAPI
	if err := json.Unmarshal(b, &UserAPI); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	token, err := h.auth.RegisterUser(UserAPI, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("JWT " + token))
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var UserAPI models.UserAPI
	if err := json.Unmarshal(b, &UserAPI); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := h.auth.AuthUser(UserAPI, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "Bearer",
		Value: token,
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("JWT " + token))
}

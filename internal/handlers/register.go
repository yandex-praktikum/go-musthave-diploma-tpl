package handlers

import (
	"encoding/json"
	"net/http"
)

type userRegisterReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h HTTPHandlers) registerUser(writer http.ResponseWriter, request *http.Request) {
	var req userRegisterReq
	err := json.NewDecoder(request.Body).Decode(&req)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(req.Password) == 0 || len(req.Login) == 0 {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := h.userService.RegisterUser(request.Context(), req.Login, req.Password)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	authToken, err := h.jwtClient.BuildJWTString(id)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	setAuthCookie(writer, authToken, h.jwtClient.GetTokenExp())
	writer.WriteHeader(http.StatusOK)
	return
}

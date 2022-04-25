package handlers

import (
	"errors"
	"net/http"

	"github.com/abayken/yandex-practicum-diploma/internal/custom_errors"
	"github.com/abayken/yandex-practicum-diploma/internal/usecases"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	AuthUseCase usecases.AuthUseCase
}

func (handler *Handler) RegisterUser(ctx *gin.Context) {
	type Request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	var request Request

	if err := ctx.BindJSON(&request); err != nil {
		ctx.Status(http.StatusBadRequest)

		return
	}

	err := handler.AuthUseCase.Register(request.Login, request.Password)

	if err != nil {
		var userExistsError *custom_errors.AlreadyExistsUserError
		if errors.As(err, &userExistsError) {
			ctx.Status(http.StatusConflict)
		} else {
			ctx.Status(http.StatusInternalServerError)
		}

		return
	}

	ctx.String(http.StatusAccepted, "hello")
}

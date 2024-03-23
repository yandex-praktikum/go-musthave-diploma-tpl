package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/A-Kuklin/gophermart/api/types"
	"github.com/A-Kuklin/gophermart/internal/auth"
	"github.com/A-Kuklin/gophermart/internal/usecases"
)

func (a *API) CreateUser(gCtx *gin.Context) {
	req := types.UserCreateRequest{}

	if err := gCtx.ShouldBindJSON(&req); err != nil {
		a.logger.WithError(err).Info("failed to bind create user request")
		gCtx.AbortWithError(http.StatusBadRequest, err)

		return
	}

	createUseCaseArgs := req.ToDomain()

	user, err := a.Usecases.User.Create(gCtx, createUseCaseArgs)
	var pgErr *pgconn.PgError
	switch {
	case errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation:
		resp := new(types.ResponseMeta).FromMessage("login", "Login already exists")
		gCtx.AbortWithStatusJSON(http.StatusConflict, resp)
		return
	case err != nil:
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	token, err := auth.Auth.GetToken(user.ID)
	if err != nil {
		resp := new(types.ResponseMeta).FromMessage("token", "Signing token error")
		gCtx.AbortWithStatusJSON(http.StatusInternalServerError, resp)
		return
	}

	gCtx.Writer.Header().Set(auth.AccessTokenHeaderKey, token)
	gCtx.JSON(http.StatusOK, types.UserCreateResponse{User: user, Token: token})
}

func (a *API) Login(gCtx *gin.Context) {
	req := types.UserCreateRequest{}

	if err := gCtx.ShouldBindJSON(&req); err != nil {
		a.logger.WithError(err).Info("failed to bind create user request")
		gCtx.AbortWithError(http.StatusBadRequest, err)

		return
	}

	createUseCaseArgs := req.ToDomain()
	user, err := a.Usecases.User.Login(gCtx, createUseCaseArgs)
	switch {
	case errors.As(err, &usecases.ErrInvalidPassword):
		resp := new(types.ResponseMeta).FromMessage("password", "Invalid password")
		gCtx.AbortWithStatusJSON(http.StatusUnauthorized, resp)
		return
	case errors.As(err, &pgx.ErrNoRows):
		resp := new(types.ResponseMeta).FromMessage("login", "Invalid login")
		gCtx.AbortWithStatusJSON(http.StatusUnauthorized, resp)
		return
	case err != nil:
		resp := new(types.ResponseMeta).FromMessage("login", "Server error")
		gCtx.AbortWithStatusJSON(http.StatusInternalServerError, resp)
		return
	}

	token, err := auth.Auth.GetToken(user.ID)
	if err != nil {
		resp := new(types.ResponseMeta).FromMessage("token", "Signing token error")
		gCtx.AbortWithStatusJSON(http.StatusInternalServerError, resp)
		return
	}

	gCtx.Writer.Header().Set(auth.AccessTokenHeaderKey, token)
	gCtx.JSON(http.StatusOK, types.LoginResponse{Token: token})
}

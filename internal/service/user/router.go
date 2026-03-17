package user

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func (u *User) RegistryRoutes(api huma.API) {

	// Handler /api/user/register
	huma.Register(api, huma.Operation{
		OperationID: "user-register",
		Method:      http.MethodPost,
		Path:        "/api/user/register",
		Summary:     "Регистрация нового пользователя",
		Description: "",
		Errors: []int{
			http.StatusBadRequest,          // 400
			http.StatusConflict,            // 409
			http.StatusInternalServerError, // 500
		},
		Tags: []string{"User"},
	}, u.signUpHandler())

	// Handler /api/user/login
	huma.Register(api, huma.Operation{
		OperationID: "user-login",
		Method:      http.MethodPost,
		Path:        "/api/user/login",
		Summary:     "Авторизация пользователя",
		Description: "",
		Errors: []int{
			http.StatusBadRequest,          // 400
			http.StatusUnauthorized,        // 401
			http.StatusInternalServerError, // 500
		},
		Tags: []string{"User"},
	}, u.signInHandler())
}

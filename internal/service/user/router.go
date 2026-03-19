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
		Responses: map[string]*huma.Response{
			"200": {Description: "Успешный ответ в формате JSON"},
			"400": {
				Description: "Неверный формат запроса",
				Content: map[string]*huma.MediaType{
					"application/json": {
						Schema: &huma.Schema{
							Ref: "#/components/schemas/ErrorModel",
						},
					},
				},
			},
			"409": {Description: "Логин уже занят"},
			"500": {Description: "Внутренняя ошибка сервера"},
		},
		Tags: []string{"User"},
	}, u.signUpHandler())

	// Handler /api/user/login
	huma.Register(api, huma.Operation{
		OperationID: "user-login",
		Method:      http.MethodPost,
		Path:        "/api/user/login",
		Summary:     "Аутентификация пользователя",
		Description: "",
		Errors: []int{
			http.StatusBadRequest,          // 400
			http.StatusUnauthorized,        // 401
			http.StatusInternalServerError, // 500
		},
		Tags: []string{"User"},
	}, u.signInHandler())
}

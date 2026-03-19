package orders

import (
	"gmart/internal/model/auth"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func (o *Orders) RegistryRoutes(api huma.API, tokenVerifier *auth.TokenVerifier) {

	// Handler POST /api/user/orders
	huma.Register(api, huma.Operation{
		OperationID: "user-orders-upload",
		Method:      http.MethodPost,
		Path:        "/api/user/orders",
		Summary:     "Загрузка нового заказа",
		Responses: map[string]*huma.Response{
			"200": {Description: "Номер заказа уже был загружен этим пользователем"},
			"202": {Description: "Новый номер заказа принят в обработку"},
		},
		Errors: []int{
			http.StatusBadRequest,          // 400 — неверный формат запроса;
			http.StatusUnauthorized,        // 401 — пользователь не аутентифицирован;
			http.StatusConflict,            // 409 — номер заказа уже был загружен другим пользователем;
			http.StatusUnprocessableEntity, // 422 — неверный формат номера заказа;
			http.StatusInternalServerError, // 500 — внутренняя ошибка сервера.
		},
		Tags: []string{"Orders"},
		Middlewares: huma.Middlewares{
			auth.NewTokenVerifierMiddleware(api, tokenVerifier),
		},
		Security:      []map[string][]string{{"bearer": {}}},
		DefaultStatus: 200,
	}, o.ordersUploadHandler())

	// Handler GET /api/user/orders
	huma.Register(api, huma.Operation{
		OperationID: "user-orders-list",
		Method:      http.MethodGet,
		Path:        "/api/user/orders",
		Summary:     "Список заказов",
		Responses: map[string]*huma.Response{
			"200": {Description: "Успешная обработка запроса"},
			"204": {Description: "Нет данных для ответа"},
		},
		Errors: []int{
			http.StatusInternalServerError, // 500 — внутренняя ошибка сервера.
		},
		Tags: []string{"Orders"},
		Middlewares: huma.Middlewares{
			auth.NewTokenVerifierMiddleware(api, tokenVerifier),
		},
		Security: []map[string][]string{{"bearer": {}}},
	}, o.ordersListHandler())
}

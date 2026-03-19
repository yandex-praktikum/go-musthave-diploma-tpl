package loyalty

import (
	"gmart/internal/model/auth"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func (l *Loyalty) RegistryRoutes(api huma.API, vrf *auth.TokenVerifier) {
	authMW := auth.NewTokenVerifierMiddleware(api, vrf)

	// GET /api/user/balance
	huma.Register(api, huma.Operation{
		OperationID: "get-user-balance",
		Method:      http.MethodGet,
		Path:        "/api/user/balance",
		Summary:     "Получение текущего баланса пользователя",
		Security:    []map[string][]string{{"bearer": {}}},
		Middlewares: huma.Middlewares{authMW},
		Tags:        []string{"Loyalty"},
		Responses: map[string]*huma.Response{
			"200": {Description: "Успешная обработка запроса"},
		},
		Errors: []int{
			http.StatusUnauthorized,        // 401 — пользователь не аутентифицирован;
			http.StatusInternalServerError, // 500 — внутренняя ошибка сервера.
		},
	}, l.getBalanceHandler())

	// POST /api/user/balance/withdraw
	huma.Register(api, huma.Operation{
		OperationID: "withdraw-points",
		Method:      http.MethodPost,
		Path:        "/api/user/balance/withdraw",
		Summary:     "Запрос на списание средств",
		Security:    []map[string][]string{{"bearer": {}}},
		Middlewares: huma.Middlewares{authMW},
		Tags:        []string{"Loyalty"},
		Responses: map[string]*huma.Response{
			"200": {Description: "Успешная обработка запроса"},
		},
		Errors: []int{
			http.StatusUnauthorized,        // 401 — пользователь не аутентифицирован;
			http.StatusPaymentRequired,     // 402 — на счету недостаточно средств;
			http.StatusUnprocessableEntity, // 422 — неверный формат номера заказа;
			http.StatusInternalServerError, // 500 — внутренняя ошибка сервера.
		},
	}, l.withdrawHandler())

	// GET /api/user/withdrawals
	huma.Register(api, huma.Operation{
		OperationID: "get-withdrawals",
		Method:      http.MethodGet,
		Path:        "/api/user/withdrawals",
		Summary:     "Получение информации о выводе средств",
		Security:    []map[string][]string{{"bearer": {}}},
		Middlewares: huma.Middlewares{authMW},
		Tags:        []string{"Loyalty"},
		Responses: map[string]*huma.Response{
			"200": {Description: "Успешная обработка запроса"},
			"204": {Description: "Нет ни одного списания"},
		},
		Errors: []int{
			http.StatusUnauthorized,        // 401 — пользователь не аутентифицирован;
			http.StatusInternalServerError, // 500 — внутренняя ошибка сервера.
		},
	}, l.getWithdrawalsHandler())
}

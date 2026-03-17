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
		Security:    []map[string][]string{{"bearer": {}}},
		Middlewares: huma.Middlewares{authMW},
	}, l.getBalanceHandler())

	// POST /api/user/balance/withdraw
	huma.Register(api, huma.Operation{
		OperationID: "withdraw-points",
		Method:      http.MethodPost,
		Path:        "/api/user/balance/withdraw",
		Security:    []map[string][]string{{"bearer": {}}},
		Middlewares: huma.Middlewares{authMW},
	}, l.withdrawHandler())

	// GET /api/user/withdrawals
	huma.Register(api, huma.Operation{
		OperationID: "get-withdrawals",
		Method:      http.MethodGet,
		Path:        "/api/user/withdrawals",
		Security:    []map[string][]string{{"bearer": {}}},
		Middlewares: huma.Middlewares{authMW},
	}, l.getWithdrawalsHandler())
}

# handler

HTTP layer of gophermart: `apiHandler` + chi router.

- `router.go` — route table: public `/register`, `/login`; token-protected orders, balance and withdrawal endpoints.
- `core.go` — `apiHandler` struct and the shared `decodeJSON` helper.
- `handler_user.go`, `handler_orders.go`, `handler_balance.go`, `handler_withdraw.go` — one method per endpoint; decode JSON, map service errors to HTTP statuses.
- `middleware.go` — `requireJSONContent`, `validateToken` (JWT from the `Authorization` header), username context key.

Contains no business logic; everything is delegated to `service.LoyaltyService`.

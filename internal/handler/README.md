# handler

HTTP layer of gophermart: `apiHandler` + chi router.

- `router.go` — route table: public `/register`, `/login`; token-protected orders, balance and withdrawal endpoints.
- `handler.go` — one method per endpoint; decodes JSON, maps service errors to HTTP statuses.
- `middleware.go` — `requireJSONContent`, `validateToken` (JWT from the `Authorization` header), username context key.

Contains no business logic; everything is delegated to `service.LoyaltyService`.

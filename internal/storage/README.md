# storage

PostgreSQL implementation of the user/order queries. `PostgresStorage` structurally satisfies the `Repository` interfaces declared in `service` and `accrual` (wired in `cmd/gophermart`; this package imports neither).

- `core.go` — `PostgresStorage` struct, constructor and the `withRetries` backoff helper.
- `service_user.go` — `service.Repository` user methods: existence check, insert, password hash lookup.
- `service_order.go` — `service.Repository` order methods: order insert with owner detection, order listing.
- `service_withdrawal.go` — `service.Repository` balance and withdrawal methods: balance, withdraw, withdrawal history.
- `accrual.go` — `accrual.Repository` methods: fetching orders to poll, persisting accrual updates.
- `templates.go` — SQL query templates built from table-name constants.
- `pgerrors.go` — classifies PostgreSQL error codes as retriable / non-retriable (used by `withRetries`).

Knows nothing about HTTP or business rules; consumes `model` types only.

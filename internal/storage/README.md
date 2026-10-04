# storage

PostgreSQL implementation of the user/order queries. `PostgresStorage` structurally satisfies the `Repository` interfaces declared in `service` and `accrual` (wired in `cmd/gophermart`; this package imports neither).

- `postgres.go` — user and order queries: registration, order insert with owner detection, accrual status updates.
- `templates.go` — SQL query templates built from table-name constants.
- `pgerrors.go` — classifies PostgreSQL error codes as retriable / non-retriable (used by `withRetries`).

Knows nothing about HTTP or business rules; consumes `model` types only.

# accrual

Background integration with the loyalty accrual system.

- `worker.go` — `accrualWorker.ProcessNewOrders`: polls on a ticker, fetches NEW/PROCESSING orders, updates their status.
- `client.go` — `accrualClient.GetStatus`: HTTP GET of the order status; maps responses to sentinel errors (204, 429, 4xx, 5xx).
- `storage.go` — `Repository` interface consumed by the worker; it is satisfied structurally by `*storage.PostgresStorage`, which is wired in `cmd/gophermart`. The package itself imports nothing from `storage`.
- `model.go` / `errors.go` — response payload and error contract.

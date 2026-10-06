# model

Shared domain types and errors used by all layers.

- `User` — request struct with a `Validate` method; `order_number.go` — `ValidateOrderNumber` (empty check + Luhn).
- `OrderResponse`, `BalanceResponse`, `WithdrawalRequest`, `WithdrawalResponse` — API payloads with JSON tags.
- `OrderRegistrationResult`, `WithdrawalRegistrationResult` — insert results with owner detection.
- Order lifecycle statuses (`NEW`, `PROCESSING`, `INVALID`, `PROCESSED`, `REGISTERED`) — untyped string constants.
- Sentinel errors (`Err*`) — the contract between storage, service and handler layers.

The package has no dependencies on other internal packages.

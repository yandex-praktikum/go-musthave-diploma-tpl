# model

Shared domain types and errors used by all layers.

- `User`, `Order` — request/domain structs with `Validate` methods (username/password, order number via Luhn check).
- `OrderResponse`, `BalanceResponse`, `WithdrawalRequest`, `WithdrawalResponse` — API payloads with JSON tags.
- `Status` — order lifecycle states (`NEW`, `PROCESSING`, `INVALID`, `PROCESSED`, `REGISTERED`).
- Sentinel errors (`Err*`) — the contract between storage, service and handler layers.

The package has no dependencies on other internal packages.

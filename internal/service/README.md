# service

Business logic of gophermart: `LoyaltyService`.

- User registration and login (bcrypt hashing, token issuing).
- Order loading with owner detection (`OrderWriteResult`).
- Balance, withdrawals and order history (partly TODO).

Defines the interfaces it consumes (satisfied structurally, wired in `cmd/gophermart`):

- `Repository` (`storage.go`) — implemented by `*storage.PostgresStorage`.
- `AuthManager` (`auth.go`) — implemented by `*auth.JWTManager`.

HTTP-agnostic: called by handlers, wrapped with `context` timeouts.

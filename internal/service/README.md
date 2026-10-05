# service

Business logic of gophermart: `LoyaltyService`.

- User registration and login (bcrypt hashing, token issuing).
- Order loading with owner detection (`OrderRegistrationResult`).
- Balance, withdrawals (duplicate detection via `WithdrawalRegistrationResult`) and order history.

Defines the interfaces it consumes (satisfied structurally, wired in `cmd/gophermart`):

- `Repository` (`storage.go`) — implemented by `*storage.PostgresStorage`.
- `AuthManager` (`auth.go`) — implemented by `*auth.JWTManager`.

HTTP-agnostic: called by handlers, wrapped with `context` timeouts.

## Design note

`LoyaltyService` is a classic facade-orchestrator: a single point where every
use case follows the same flow — validate → call the repository or auth
manager → audit-log. Splitting it by responsibility (a validator, a
repository proxy, a logger) would smear one flow across files; keeping it in
one place keeps each use case readable top-to-bottom.

# auth

JWT issuing and verification: `JWTManager`.

- Signs HMAC (HS256) tokens with a secret key; TTL is 3 hours.
- `Claims` embeds registered claims plus the username.
- Implements the `service.AuthManager` interface.

No database or HTTP dependencies.

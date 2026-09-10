# Authentication

WebSocket upgrades require an `Authorization: Bearer <JWT>` header. The server accepts only HS256,
checks issuer and expiry, requires a subject, and permits five seconds of clock skew. A secret shorter
than 32 bytes prevents startup.

The event sender is built from signed `sub` and `display_name` claims. Client payload fields cannot
replace authenticated identity.

The `cmd/token` utility exists only for local demonstration with synthetic users; it is not a login
service. The later browser client will use a short-lived ticket exchange instead of tokens in URLs.

Production deployments should prefer asymmetric signing or a managed identity provider, rotate
keys, validate audience, support revocation, rate-limit authentication, and never log tokens.

# Go Realtime Messaging Platform

A scalable real-time messaging portfolio project built with Go and WebSocket, focused on concurrency
safety, authenticated protocol boundaries, explicit backpressure, and horizontal scalability.

## Milestone v0.2.0

- HS256 JWT issuance and strict verification
- authenticated WebSocket handshake through the Authorization header
- sender identity derived from signed claims, never client payloads
- issuer, expiry, algorithm, subject, and minimum-secret validation
- development CLI for short-lived synthetic user tokens
- unauthorized and identity-spoofing integration tests
- race-enabled CI, graceful shutdown, bounded queues, and message limits

## Run

```bash
export JWT_SECRET='replace-this-with-a-long-random-development-secret'
go run ./cmd/server
```

Create a short-lived synthetic token in another terminal:

```bash
export JWT_SECRET='replace-this-with-a-long-random-development-secret'
TOKEN=$(go run ./cmd/token -user demo-alice -name Alice)
```

Connect to `ws://localhost:8080/ws` with `Authorization: Bearer $TOKEN`. The server-supplied
`sender.id` always comes from the verified token.

## Architecture

See [architecture](docs/ARCHITECTURE.md) and [authentication](docs/AUTHENTICATION.md).

## Roadmap

1. PostgreSQL-backed users, rooms, memberships, and message history
2. private and group rooms with presence, typing, and read receipts
3. Redis Pub/Sub for horizontal scaling
4. Next.js client with a browser-safe WebSocket authentication exchange
5. load testing, metrics, Docker Compose, and v1.0.0 release

## License

MIT

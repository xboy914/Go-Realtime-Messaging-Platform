# Go Realtime Messaging Platform

A scalable real-time messaging portfolio project built with Go, WebSocket, PostgreSQL, and JWT.

## Milestone v0.3.0

- versioned, embedded PostgreSQL migrations
- users, rooms, memberships, and indexed message history
- transactional membership checks before message persistence
- room-scoped WebSocket fan-out that excludes non-members
- authenticated REST history endpoint
- synthetic Alice, Bob, outsider, and demo-room fixtures
- PostgreSQL integration tests under the race detector

## Run

```bash
export JWT_SECRET='replace-this-with-a-long-random-development-secret'
export DATABASE_URL='postgres://messaging:password@localhost:5432/messaging?sslmode=disable'
go run ./cmd/server
```

Create a short-lived synthetic token:

```bash
TOKEN=$(go run ./cmd/token -user demo-alice -name Alice)
```

Use `Authorization: Bearer $TOKEN` for `GET /rooms/demo-room/messages` and the `/ws`
handshake. Messages are persisted before broadcast and delivered only to connected room members.

## Architecture

See [architecture](docs/ARCHITECTURE.md) and [authentication](docs/AUTHENTICATION.md).

## Roadmap

1. presence, typing, and read receipts
2. Redis Pub/Sub for horizontal scaling
3. Next.js client with browser-safe WebSocket authentication
4. load testing, metrics, Docker Compose, and v1.0.0 release

## License

MIT

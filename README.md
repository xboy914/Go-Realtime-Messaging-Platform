# Go Realtime Messaging Platform

A scalable real-time messaging portfolio project built with Go and WebSocket. The implementation
focuses on concurrency safety, explicit backpressure, testable protocol boundaries, and gradual
evolution toward a multi-node messaging system.

## Milestone v0.1.0

- HTTP server with graceful shutdown and defensive timeouts
- WebSocket endpoint at `/ws`
- channel-driven hub with one owner goroutine
- bounded client queues and slow-consumer eviction
- validated JSON event envelope and 16 KiB message limit
- health endpoint and security headers
- race-enabled automated tests and CI

## Run

```bash
go mod download
go run ./cmd/server
```

Health check:

```bash
curl http://localhost:8080/healthz
```

Connect a WebSocket client to `ws://localhost:8080/ws` and send:

```json
{"type":"message.created","room_id":"demo-room","payload":{"text":"hello"}}
```

## Architecture

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for the concurrency model and event contract.

## Roadmap

1. JWT authentication and user identity
2. PostgreSQL-backed users, rooms, memberships, and message history
3. private and group rooms with presence, typing, and read receipts
4. Redis Pub/Sub for horizontal scaling
5. Next.js client
6. load testing, metrics, Docker Compose, and v1.0.0 release

## License

MIT

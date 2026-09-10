.PHONY: run test check

run:
	go run ./cmd/server

test:
	go test -race ./...

check:
	gofmt -w .
	go vet ./...
	go test -race ./...
	go build ./cmd/server

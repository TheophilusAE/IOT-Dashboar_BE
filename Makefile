.PHONY: run build test vet verify-ws

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

vet:
	go vet ./...

# Starts a WebSocket listener to watch real-time pushes while you run the
# scripts/verify/send_*.sh scripts in another terminal.
verify-ws:
	go run ./scripts/verify/ws_listen.go

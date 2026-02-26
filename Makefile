.PHONY: run build test lint tidy sqlc

# Run the server in development mode
run:
	go run cmd/server/main.go

# Build the server binary
build:
	go build -o bin/server cmd/server/main.go

# Run all tests with race detection
test:
	go test -v -race ./...

# Run linter
lint:
	golangci-lint run

# Tidy dependencies
tidy:
	go mod tidy

# Generate sqlc code
sqlc:
	sqlc generate

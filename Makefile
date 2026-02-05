.PHONY: build run test clean deps help

# Build the server
build:
	go build -o dbhub-mcp-server ./cmd/server

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o dbhub-mcp-server-linux ./cmd/server
	GOOS=windows GOARCH=amd64 go build -o dbhub-mcp-server.exe ./cmd/server
	GOOS=darwin GOARCH=arm64 go build -o dbhub-mcp-server-mac ./cmd/server

# Run the server (requires .env file)
run: build
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Copy .env.example to .env and configure it."; \
		exit 1; \
	fi
	@export $$(cat .env | xargs) && ./dbhub-mcp-server

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run tests with race detection
test-race:
	go test -race ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	rm -f dbhub-mcp-server dbhub-mcp-server.exe dbhub-mcp-server-*
	rm -f coverage.out coverage.html
	rm -f *.test

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	go vet ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the server"
	@echo "  build-all     - Build for Linux, Windows, and macOS"
	@echo "  run           - Build and run the server (requires .env)"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-race     - Run tests with race detection"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  clean         - Remove build artifacts"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  help          - Show this help"

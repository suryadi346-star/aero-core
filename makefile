.PHONY: help build run vet lint test docker-up docker-down clean

BIN_NAME := aero-core
GO_FLAGS := CGO_ENABLED=0 GOOS=linux
LDFLAGS := -ldflags="-s -w"

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build static binary
	$(GO_FLAGS) go build $(LDFLAGS) -o $(BIN_NAME) ./cmd/server

run: ## Run server locally
	go run ./cmd/server

vet: ## Run static analysis
	go vet ./...

lint: ## Run linter (requires golangci-lint)
	golangci-lint run ./...

test: ## Run tests
	go test -race -coverprofile=coverage.out ./...

docker-up: ## Start services via Docker Compose
	docker compose up -d

docker-down: ## Stop & remove containers
	docker compose down --remove-orphans

clean: ## Remove build artifacts
	rm -f $(BIN_NAME) coverage.out

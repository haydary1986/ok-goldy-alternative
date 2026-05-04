SHELL := /bin/bash

APP_NAME := ok-goldy-alternative
GO       := go
DOCKER   := docker
COMPOSE  := docker compose
MIGRATE  := $(GO) run ./cmd/migrate

.PHONY: help deps tidy build run run-worker test lint fmt vet clean docker-build docker-up docker-down migrate-up migrate-down migrate-status web-install web-dev web-build web-typecheck

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

deps: ## Resolve Go module dependencies
	$(GO) mod tidy

tidy: deps ## Alias for deps

build: ## Build server, worker, and migrate binaries
	mkdir -p bin
	$(GO) build -o bin/server  ./cmd/server
	$(GO) build -o bin/worker  ./cmd/worker
	$(GO) build -o bin/migrate ./cmd/migrate

run: ## Run the API server locally
	$(GO) run ./cmd/server

run-worker: ## Run the background worker locally
	$(GO) run ./cmd/worker

test: ## Run unit tests
	$(GO) test -race -cover ./...

lint: ## Run go vet
	$(GO) vet ./...

fmt: ## Format all Go code
	$(GO) fmt ./...

clean: ## Remove build artifacts
	rm -rf bin dist tmp coverage.out coverage.html

docker-build: ## Build the Goldy Docker image
	$(DOCKER) build -t $(APP_NAME):latest .

docker-up: ## Bring up the full local stack
	$(COMPOSE) up -d --build

docker-down: ## Stop the local stack
	$(COMPOSE) down

docker-logs: ## Tail logs of the running stack
	$(COMPOSE) logs -f --tail=200

migrate-up: ## Apply pending migrations
	$(MIGRATE) up

migrate-down: ## Roll back the most recent migration
	$(MIGRATE) down

migrate-status: ## Show migration status
	$(MIGRATE) status

# --- Web frontend ---

web-install: ## Install frontend dependencies
	cd web && npm install

web-dev: ## Run the SPA dev server (Vite, port 5173, proxies /api -> :8080)
	cd web && npm run dev

web-build: ## Build the SPA for production (web/dist)
	cd web && npm run build

web-typecheck: ## Type-check the frontend without emitting
	cd web && npm run typecheck

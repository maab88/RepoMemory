COMPOSE_FILE=docker-compose.yml
GOLANGCI_LINT ?= go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

.PHONY: up down logs test test-web test-api test-worker lint lint-web lint-go format format-web format-go ci dev-web dev-api dev-worker generate-contracts

up:
	docker compose --env-file .env up -d

down:
	docker compose --env-file .env down

logs:
	docker compose --env-file .env logs -f

test: test-api test-worker test-web

test-web:
	corepack pnpm --dir apps/web test

test-api:
	cd apps/api && go test ./...

test-worker:
	cd apps/worker && go test ./...

lint: lint-go lint-web

lint-web:
	corepack pnpm --dir apps/web lint

lint-go:
	cd apps/api && $(GOLANGCI_LINT) run
	cd apps/worker && $(GOLANGCI_LINT) run

format: format-go format-web

format-web:
	corepack pnpm --dir apps/web format

format-go:
	cd apps/api && gofmt -w ./cmd ./internal
	cd apps/worker && gofmt -w ./cmd ./internal

ci: lint test

dev-web:
	corepack pnpm --dir apps/web dev

dev-api:
	cd apps/api && go run ./cmd/api

dev-worker:
	cd apps/worker && go run ./cmd/worker

generate-contracts:
	corepack pnpm --dir packages/contracts generate

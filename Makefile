COMPOSE_FILE=docker-compose.yml

.PHONY: up down logs test lint dev-web dev-api dev-worker

up:
	docker compose --env-file .env up -d

down:
	docker compose --env-file .env down

logs:
	docker compose --env-file .env logs -f

test:
	cd apps/api && go test ./...
	pnpm --dir apps/web test

lint:
	cd apps/api && go test ./...
	pnpm --dir apps/web lint

dev-web:
	pnpm --dir apps/web dev

dev-api:
	cd apps/api && go run ./cmd/api

dev-worker:
	cd apps/worker && go run ./cmd/worker
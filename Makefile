.PHONY: dev-web dev-api dev-worker docker-up docker-down migrate seed test-web test-server test-e2e lint build

# ---- Development ----

dev-web:
	cd apps/web && npm run dev

dev-api:
	cd apps/server && go run cmd/api/main.go

dev-worker:
	cd apps/server && go run cmd/judge-worker/main.go

# ---- Docker Infrastructure ----

docker-up:
	docker compose -f infra/docker-compose.yml up -d

docker-down:
	docker compose -f infra/docker-compose.yml down

# ---- Database ----

migrate:
	cd apps/server && go run cmd/migrate/main.go

seed:
	cd apps/server && go run cmd/seed/main.go

# ---- Testing ----

test-web:
	cd apps/web && npm run test

test-server:
	cd apps/server && go test ./...

test-e2e:
	cd apps/web && npm run test:e2e

# ---- Code Quality ----

lint:
	cd apps/web && npm run lint
	cd apps/server && golangci-lint run ./...

# ---- Build ----

build:
	cd apps/web && npm run build
	cd apps/server && go build -o bin/api ./cmd/api

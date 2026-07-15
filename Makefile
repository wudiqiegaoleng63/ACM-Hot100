.PHONY: dev-web dev-api dev-worker docker-up docker-down migrate seed test-web test-server test-e2e lint build verify verify-web verify-server verify-compose

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
	cd apps/server && go run cmd/api/main.go -migrate

seed:
	cd apps/server && go run cmd/api/main.go -seed

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
	cd apps/server && go vet ./...

# ---- Build ----

build:
	cd apps/web && npm run build
	cd apps/server && go build ./cmd/api ./cmd/judge-worker

# ---- Verification ----

verify: verify-web verify-server verify-compose

verify-web:
	cd apps/web && npm run lint
	cd apps/web && npm run type-check
	cd apps/web && npm run test -- --run
	cd apps/web && npm run build

verify-server:
	cd apps/server && go test ./...
	cd apps/server && go vet ./...
	cd apps/server && go build ./cmd/api ./cmd/judge-worker

verify-compose:
	docker compose -f infra/docker-compose.yml config --quiet

.PHONY: run build test fmt vet tidy db-up db-reset migrate seed

GO ?= go
BIN := bin/api

run:
	$(GO) run ./cmd/api

build:
	$(GO) build -o $(BIN) ./cmd/api

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

db-up:
	docker compose up -d --wait

db-reset:
	./scripts/setup-db.sh

migrate:
	@. ./.env && migrate -path migrations -database "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$${POSTGRES_HOST:-localhost}:$${POSTGRES_PORT:-5432}/$$POSTGRES_DB" up

seed:
	@. ./.env && psql "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$${POSTGRES_HOST:-localhost}:$${POSTGRES_PORT:-5432}/$$POSTGRES_DB" -f seeds/seed.sql

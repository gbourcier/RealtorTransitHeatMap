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
	@. ./.env && migrate -path migrations -database "$$DATABASE_URL" up

seed:
	@. ./.env && psql "$$DATABASE_URL" -f seeds/seed.sql

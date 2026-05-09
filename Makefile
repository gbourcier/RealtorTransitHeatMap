.PHONY: run dev build build-backend test fmt vet tidy db-up db-reset migrate seed swagger

GO ?= go
BIN := bin/api

run:
	$(GO) run ./cmd/api

dev:
	$(GO) build -gcflags='all=-N -l' -o ./tmp/api ./cmd/api && dlv exec ./tmp/api --listen=127.0.0.1:2345 --headless=true --api-version=2 --accept-multiclient --continue --log --

swagger:
	swag init -g cmd/api/main.go

build-backend: swagger
	$(GO) build -o $(BIN) ./cmd/api

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

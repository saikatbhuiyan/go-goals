ifneq (,$(wildcard .env))
include .env
export
endif

GOOSE ?= go run github.com/pressly/goose/v3/cmd/goose@latest
MIGRATIONS_DIR ?= migrations
DB_DRIVER ?= postgres
MIGRATION_NAME ?= change

.PHONY: dev-api dev-web test-api lint-web build-web goose-install migrate-up migrate-down migrate-redo migrate-reset migrate-status migrate-version migrate-create

dev-api:
	go run ./apps/api/cmd/server

dev-web:
	npm --prefix apps/web run dev

test-api:
	go test ./apps/api/...

lint-web:
	npm --prefix apps/web run lint

build-web:
	npm --prefix apps/web run build

goose-install:
	go install github.com/pressly/goose/v3/cmd/goose@latest

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(POSTGRES_URL)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(POSTGRES_URL)" down

migrate-redo:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(POSTGRES_URL)" redo

migrate-reset:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(POSTGRES_URL)" reset

migrate-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(POSTGRES_URL)" status

migrate-version:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(POSTGRES_URL)" version

migrate-create:
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(MIGRATION_NAME) sql

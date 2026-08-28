ifneq (,$(wildcard .env))
include .env
export
endif

GOOSE ?= go run github.com/pressly/goose/v3/cmd/goose@latest
API_DIR ?= apps/api
MIGRATIONS_DIR ?= migrations
DB_DRIVER ?= postgres
MIGRATION_NAME ?= change

.PHONY: dev-api dev-web test-api lint-web build-web goose-install migrate-up migrate-down migrate-redo migrate-reset migrate-status migrate-version migrate-create

dev-api:
	cd $(API_DIR) && go run ./cmd/server

dev-web:
	npm --prefix apps/web run dev

test-api:
	cd $(API_DIR) && go test ./...

lint-web:
	npm --prefix apps/web run lint

build-web:
	npm --prefix apps/web run build

goose-install:
	go install github.com/pressly/goose/v3/cmd/goose@latest

migrate-up:
	cd $(API_DIR) && $(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(POSTGRES_URL)" up

migrate-down:
	cd $(API_DIR) && $(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(POSTGRES_URL)" down

migrate-redo:
	cd $(API_DIR) && $(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(POSTGRES_URL)" redo

migrate-reset:
	cd $(API_DIR) && $(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(POSTGRES_URL)" reset

migrate-status:
	cd $(API_DIR) && $(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(POSTGRES_URL)" status

migrate-version:
	cd $(API_DIR) && $(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(POSTGRES_URL)" version

migrate-create:
	cd $(API_DIR) && $(GOOSE) -dir $(MIGRATIONS_DIR) create $(MIGRATION_NAME) sql

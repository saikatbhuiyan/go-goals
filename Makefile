ENV_FILE ?= .env

ifneq (,$(wildcard $(ENV_FILE)))
include $(ENV_FILE)
export
endif

GOOSE ?= go run github.com/pressly/goose/v3/cmd/goose@latest
API_DIR ?= apps/api
MIGRATIONS_DIR ?= migrations
DB_DRIVER ?= postgres
MIGRATION_NAME ?= change

.PHONY: dev-api prod-api dev-web test-api lint-web build-web goose-install migrate-up migrate-down migrate-redo migrate-reset migrate-status migrate-version migrate-create prod-migrate-up prod-migrate-down prod-migrate-redo prod-migrate-reset prod-migrate-status prod-migrate-version

dev-api:
	cd $(API_DIR) && go run ./cmd/server

prod-api:
	$(MAKE) ENV_FILE=.env.prod dev-api

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

prod-migrate-up:
	$(MAKE) ENV_FILE=.env.prod migrate-up

prod-migrate-down:
	$(MAKE) ENV_FILE=.env.prod migrate-down

prod-migrate-redo:
	$(MAKE) ENV_FILE=.env.prod migrate-redo

prod-migrate-reset:
	$(MAKE) ENV_FILE=.env.prod migrate-reset

prod-migrate-status:
	$(MAKE) ENV_FILE=.env.prod migrate-status

prod-migrate-version:
	$(MAKE) ENV_FILE=.env.prod migrate-version

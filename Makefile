ENV_FILE ?= .env

ifneq (,$(wildcard $(ENV_FILE)))
include $(ENV_FILE)
export
endif

GOOSE_VERSION ?= v3.26.0
GOOSE ?= go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)
API_DIR ?= apps/api
MIGRATIONS_DIR ?= migrations
DB_DRIVER ?= $(or $(GOOSE_DRIVER),postgres)
DB_STRING ?= $(or $(POSTGRES_URL),$(GOOSE_DBSTRING))
MIGRATION_NAME ?= change
API_BIN ?= /tmp/go-goals-api
AWS_ACCOUNT_ID ?= 109684670544
AWS_DEFAULT_REGION ?= us-east-1
API_IMAGE_REPOSITORY ?= go-goals-api
AWS_ECR_DOMAIN ?= $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_DEFAULT_REGION).amazonaws.com
BUILD_IMAGE ?= $(if $(AWS_ACCOUNT_ID),$(AWS_ECR_DOMAIN)/$(API_IMAGE_REPOSITORY),$(API_IMAGE_REPOSITORY))
BUILD_TAG ?= latest
GIT_SHA ?= $(shell git rev-parse HEAD)
DOCKER_PLATFORM ?= linux/amd64

.PHONY: dev-api prod-api dev-web build-api test-api lint-web build-web build-image build-image-login build-image-push build-image-pull build-image-promote up down goose-install require-db-string migrate-up migrate-down migrate-redo migrate-reset migrate-status migrate-version migrate-validate migrate-create prod-migrate-up prod-migrate-down prod-migrate-redo prod-migrate-reset prod-migrate-status prod-migrate-version prod-migrate-validate

dev-api:
	cd $(API_DIR) && go run ./cmd/server

prod-api:
	$(MAKE) ENV_FILE=.env.prod dev-api

dev-web:
	npm --prefix apps/web run dev

build-api:
	cd $(API_DIR) && go build -trimpath -o $(API_BIN) ./cmd/server

test-api:
	cd $(API_DIR) && go test ./...

lint-web:
	npm --prefix apps/web run lint

build-web:
	npm --prefix apps/web run build

build-image:
	docker buildx build --platform "$(DOCKER_PLATFORM)" --tag "$(BUILD_IMAGE):$(GIT_SHA)" --load $(API_DIR)

build-image-login:
	aws ecr get-login-password --region $(AWS_DEFAULT_REGION) | docker login --username AWS --password-stdin $(AWS_ECR_DOMAIN)

build-image-push: build-image-login
	docker buildx build --platform "$(DOCKER_PLATFORM)" --tag "$(BUILD_IMAGE):$(GIT_SHA)" --push $(API_DIR)

build-image-pull: build-image-login
	docker image pull "$(BUILD_IMAGE):$(GIT_SHA)"

build-image-promote: build-image-login
	docker image pull "$(BUILD_IMAGE):$(GIT_SHA)"
	docker image tag "$(BUILD_IMAGE):$(GIT_SHA)" "$(BUILD_IMAGE):$(BUILD_TAG)"
	docker image push "$(BUILD_IMAGE):$(BUILD_TAG)"

down:
	docker compose down --remove-orphans --volumes

up: down
	docker compose up --detach

goose-install:
	go install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)

require-db-string:
	@test -n "$(DB_STRING)" || (echo "DB_STRING is empty. Set POSTGRES_URL or GOOSE_DBSTRING before running migrations." >&2; exit 1)

migrate-up: require-db-string
	cd $(API_DIR) && $(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_STRING)" up

migrate-down: require-db-string
	cd $(API_DIR) && $(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_STRING)" down

migrate-redo: require-db-string
	cd $(API_DIR) && $(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_STRING)" redo

migrate-reset: require-db-string
	cd $(API_DIR) && $(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_STRING)" reset

migrate-status: require-db-string
	cd $(API_DIR) && $(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_STRING)" status

migrate-version: require-db-string
	cd $(API_DIR) && $(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_STRING)" version

migrate-validate:
	cd $(API_DIR) && $(GOOSE) -dir $(MIGRATIONS_DIR) validate

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

prod-migrate-validate:
	$(MAKE) ENV_FILE=.env.prod migrate-validate

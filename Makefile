.PHONY: dev-api dev-web test-api lint-web build-web

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

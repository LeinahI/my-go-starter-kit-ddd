DATABASE_URL ?= postgres://postgres:root@127.0.0.1:5433/ws?sslmode=disable

.PHONY: run-api test build swagger migrate-up migrate-down migrate-fresh migrate-fresh-seed migrate-rollback migrate-status migrate-create seed docker-db-up docker-db-down

run-api:
	go run ./cmd/api

test:
	go test ./...

build:
	go build -o bin/api ./cmd/api

swagger:
	swag init -g cmd/api/main.go --parseInternal --output docs

migrate-up:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate up

migrate-down:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate down

migrate-fresh:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate fresh

migrate-fresh-seed: migrate-fresh seed

migrate-rollback:
	@test -n "$(step)" || (echo 'usage: make migrate-rollback step=1' && exit 1)
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate rollback $(step)

migrate-status:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate status

migrate-create:
	@test -n "$(name)" || (echo 'usage: make migrate-create name=add_users_table' && exit 1)
	go run ./cmd/migrate create $(name)

seed:
	go run ./cmd/seed

docker-db-up:
	docker compose -f deployments/docker-compose.yml up -d db

docker-db-down:
	docker compose -f deployments/docker-compose.yml down

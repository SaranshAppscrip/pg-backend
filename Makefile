.PHONY: run build test tidy migrate docker-up docker-down

run:
	go run ./cmd/api

build:
	go build -o bin/nivas-api ./cmd/api

test:
	go test ./...

tidy:
	go mod tidy

migrate:
	psql "$$DATABASE_URL" -f migrations/001_init.sql
	psql "$$DATABASE_URL" -f migrations/002_organizations_auth.sql

docker-up:
	docker compose up -d

docker-down:
	docker compose down

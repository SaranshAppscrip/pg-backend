.PHONY: run build test tidy migrate bootstrap docker-up docker-down

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

bootstrap:
	go run ./cmd/bootstrap

docker-up:
	docker compose up -d

docker-down:
	docker compose down

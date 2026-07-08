.PHONY: run cron build test tidy migrate migrate-properties bootstrap docker-up docker-down

run:
	go run ./cmd/api

cron:
	go run ./cmd/cron

build:
	go build -o bin/nivas-api ./cmd/api
	go build -o bin/nivas-cron ./cmd/cron

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

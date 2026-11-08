.PHONY: build run migrate-up migrate-down test docker-build docker-run docker-compose-up docker-compose-down docker-compose-logs

build:
	@go build -o bin/api ./cmd/api

run: build
	@./bin/api

test:
	@go test -v ./...

migrate-up:
	@go run ./cmd/migrate up

migrate-down:
	@go run ./cmd/migrate down

# Docker helpers
docker-build:
	@docker build -t go-prod-api:latest .

docker-run:
	@docker run --rm -p 8080:8080 --env-file .env go-prod-api:latest

docker-compose-up:
	@docker compose up --build -d

docker-compose-down:
	@docker compose down

docker-compose-logs:
	@docker compose logs -f

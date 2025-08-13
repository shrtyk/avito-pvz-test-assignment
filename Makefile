.PHONY: run/main

run/main:
	@go run ./cmd/app

docker/up:
	@docker compose up --build

docker/down:
	@docker compose down --volumes

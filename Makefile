include .env

.PHONY: docker/up docker/down migrations/new migrations/up migrations/up-by-one migrations/down migrations/down-all migrations/status

MIGRATIONS_DIR=./migrations
RSA_PRIVATE=./keys/rsa/private_key.pem
RSA_PUBLIC=./keys/rsa/public_key.pem

# Run app
run/app:
	@trap 'docker-compose down pvz' EXIT; \
	docker compose up --build pvz

# Run db with migrations in background
start/db:
	@docker compose up -d --build postgres goose
	@docker compose logs goose

# Stop db
stop/db:
	@docker compose down --volumes postgres goose

# Build containers and start them in background
docker/up:
	@docker compose up -d --build

# Stop and remove containers with their volumes
docker/down:
	@docker compose down --volumes

# Recompile pvz.proto
compile/pvz-proto:
	@protoc -I ./proto/pvz \
	--go_out ./proto/pvz/gen --go_opt=paths=source_relative \
	--go-grpc_out ./proto/pvz/gen --go-grpc_opt=paths=source_relative \
	./proto/pvz/pvz.proto

# Creating migrations:
migrations/new:
	@if [ -z "$(NAME)" ]; then \
		echo "NAME is not set. Usage: make migrations/new NAME=your_migration_name"; \
		exit 1; \
	fi
	@goose -dir $(MIGRATIONS_DIR) create $(NAME) sql

# Migrate the DB to the most recent version available
migrations/up:
	@docker compose run --rm goose up

# Migrate the DB up by 1
migrations/up-by-one:
	@docker compose run --rm goose up-by-one

# Roll back the version by 1
migrations/down:
	@docker compose run --rm goose down

# Roll back all migrations
migrations/down-all:
	@docker compose run --rm goose down-to 0

# Dump the migration status for the current DB
migrations/status:
	@docker compose run --rm goose status

# Open psql in postgres container
psql/pvz:
	@docker compose exec postgres psql -U user -d pvz-db

# Generaate DTOs
generate/dto:
	@go generate ./internal/adapters/driving/http/dto/generate.go

# Generate mocks for interfaces
generate/mocks:
	@go generate ./internal/core/ports/...

# Generate all
generate:
	@make generate/dto
	@make generate/mocks

# Create keys directory and generate them
gen/rsa:
	@mkdir -p ./keys/rsa/
	@openssl genrsa -out ${RSA_PRIVATE} 2048
	@openssl rsa -pubout -in ${RSA_PRIVATE} -out ${RSA_PUBLIC}

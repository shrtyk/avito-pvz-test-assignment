.PHONY: setup docker/up docker/down app/run db/start db/stop \
        unit-tests/run integration-tests/run linter/run \
        pvz-proto/compile migrations/new migrations/up migrations/up-by-one \
        migrations/down migrations/down-all migrations/status psql/pvz \
        dto/generate mocks/generate rsa/generate generate test

MIGRATIONS_DIR=./migrations
RSA_DIR=./keys/rsa
UNIT_TESTS_PKGS := $(shell go list ./... | grep -v /mocks | grep -v /gen | grep -v /dto | grep -v /cmd)


# Setup enviroment to run
setup: rsa/generate
	@cp .env_example .env

# Build containers and start them in background
docker/up:
	@docker compose up -d --build

# Stop and remove containers with their volumes
docker/down:
	@docker compose down --volumes

# Run app
app/run:
	@trap 'docker-compose down pvz' EXIT; \
	docker compose up --build pvz

# Run db with migrations in background
db/start:
	@docker compose up -d --build postgres goose
	@docker compose logs goose

# Stop db
db/stop:
	@docker compose down --volumes postgres goose

# Run all tests
test: unit-tests/run integration-tests/run

# Run unit tests
unit-tests/run:
	@mkdir -p coverage
	@go test -v -race \
    -coverprofile=coverage/coverage.out -covermode=atomic ${UNIT_TESTS_PKGS}

# Run intgration tests
integration-tests/run:
	@go test -v -tags=integration ./cmd/app

# Run linter
linter/run:
	@golangci-lint run ./...

# Recompile pvz.proto
pvz-proto/compile:
	@protoc -I ./proto/pvz \
	--go_out ./proto/pvz/gen --go_opt=paths=source_relative \
	--go-grpc_out ./proto/pvz/gen --go-grpc_opt=paths=source_relative \
	./proto/pvz/pvz.proto

# Creating migrations
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

# Generate DTOs
dto/generate:
	@go generate ./internal/api/http/dto/generate.go

# Generate mocks for interfaces
mocks/generate:
	@go generate ./internal/core/ports/...

# Create keys directory and generate them
rsa/generate:
	@mkdir -p ${RSA_DIR}
	@openssl genrsa -out ${RSA_DIR}/private_key.pem 2048
	@openssl rsa -pubout -in ${RSA_DIR}/private_key.pem -out ${RSA_DIR}/public_key.pem

# Generate all
generate: dto/generate mocks/generate rsa/generate

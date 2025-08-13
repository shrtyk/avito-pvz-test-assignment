include .env

.PHONY: docker/up docker/down compile/pvz-proto migrations/new migrations/up migrations/up-by-one migrations/down migrations/down-all migrations/status

docker/up:
	@docker compose up -d --build

docker/down:
	@docker compose down --volumes

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
	@docker-compose run --rm goose create $(NAME) sql

# Migrate the DB to the most recent version available
migrations/up:
	@docker-compose run --rm goose up

# Migrate the DB up by 1
migrations/up-by-one:
	@docker-compose run --rm goose up-by-one

# Roll back the version by 1
migrations/down:
	@docker-compose run --rm goose down

# Roll back all migrations
migrations/down-all:
	@docker-compose run --rm goose down-to 0

# Dump the migration status for the current DB
migrations/status:
	@docker-compose run --rm goose status

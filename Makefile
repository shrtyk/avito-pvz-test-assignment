.PHONY: run/main

run/main:
	@go run ./cmd/app

docker/up:
	@docker compose up --build

docker/down:
	@docker compose down --volumes

compile/pvz-proto:
	@protoc -I ./proto/pvz \
	--go_out ./proto/pvz/gen --go_opt=paths=source_relative \
	--go-grpc_out ./proto/pvz/gen --go-grpc_opt=paths=source_relative \
	./proto/pvz/pvz.proto

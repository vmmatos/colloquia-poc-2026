PROTO_DIR  := proto
OUT_DIR    := .
BINARY     := bin/messaging

.PHONY: generate build run migrate-up migrate-down test tidy

## generate: Compile .proto files → internal/pb/ and internal/channelsclient/
generate:
	@PATH="$$PATH:$$(go env GOPATH)/bin" protoc \
		--go_out=$(OUT_DIR) \
		--go_opt=module=messaging \
		--go-grpc_out=$(OUT_DIR) \
		--go-grpc_opt=module=messaging \
		$(PROTO_DIR)/messaging.proto \
		$(PROTO_DIR)/channels_client.proto
	@echo "proto files generated in internal/pb/ and internal/channelsclient/"

## build: Build the binary
build:
	go build -o $(BINARY) ./cmd/api/main.go

## run: Build and run
run: build
	./$(BINARY)

## test: Run tests
test:
	go test ./...

## tidy: Sync go.mod / go.sum
tidy:
	go mod tidy

## migrate-up: Apply database migrations
migrate-up:
	migrate -path migrations -database "$$MESSAGING_DATABASE_URL" up

## migrate-down: Rollback last migration
migrate-down:
	migrate -path migrations -database "$$MESSAGING_DATABASE_URL" down 1

## install-tools: Install code-generation tools
install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

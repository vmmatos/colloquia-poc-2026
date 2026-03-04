PROTO_DIR  := proto
OUT_DIR    := .
BINARY     := bin/auth

.PHONY: generate build run migrate-up migrate-down test tidy

## generate: Compile .proto files → internal/pb/
generate:
	@PATH="$$PATH:$$(go env GOPATH)/bin" protoc \
		--go_out=$(OUT_DIR) \
		--go_opt=module=auth \
		--go-grpc_out=$(OUT_DIR) \
		--go-grpc_opt=module=auth \
		$(PROTO_DIR)/auth.proto
	@echo "proto files generated in internal/pb/"

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

## migrate-up: Apply database migrations (requires goose or migrate CLI)
migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

## migrate-down: Rollback last migration
migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1

## install-tools: Install code-generation tools
install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

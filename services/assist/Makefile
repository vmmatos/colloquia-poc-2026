PROTO_DIR  := proto
OUT_DIR    := .
BINARY     := bin/assist

.PHONY: generate build run test tidy install-tools

## generate: Compile .proto files → internal/pb/ and internal/messagingclient/
generate:
	@PATH="$$PATH:$$(go env GOPATH)/bin" protoc \
		--go_out=$(OUT_DIR) \
		--go_opt=module=assist \
		--go-grpc_out=$(OUT_DIR) \
		--go-grpc_opt=module=assist \
		$(PROTO_DIR)/assist.proto \
		$(PROTO_DIR)/messaging_client.proto
	@echo "proto files generated in internal/pb/ and internal/messagingclient/"

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

## install-tools: Install code-generation tools
install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

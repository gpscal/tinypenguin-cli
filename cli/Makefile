# Makefile for tinypenguin CLI

.PHONY: build build-cli build-server clean test proto-gen install

# Variables
GO_CMD = go
PROTOC_GEN_GRPC = protoc-gen-go-grpc
PROTOC_VERSION = 3.19.4
PROTO_DIR = ../proto
OUT_DIR = ./pkg
GO_MODULES = ./cmd/tinypenguin ./cmd/tinypenguin-cli

# Default target
all: build

# Generate protobuf files
proto-gen:
	@echo "Generating protobuf files..."
	@mkdir -p $(OUT_DIR)/pb
	protoc --proto_path=$(PROTO_DIR) \
		--plugin=protoc-gen-go=$(HOME)/go/bin/protoc-gen-go \
		--plugin=protoc-gen-go-grpc=$(HOME)/go/bin/protoc-gen-go-grpc \
		--go_out=$(OUT_DIR) \
		--go-grpc_out=$(OUT_DIR) \
		$(PROTO_DIR)/tinypenguin/*.proto
	@echo "Moving generated files to correct location..."
	@if [ -d "$(OUT_DIR)/example.com/tinypenguin/pkg/pb" ]; then \
		mkdir -p $(OUT_DIR)/pb; \
		cp -r $(OUT_DIR)/example.com/tinypenguin/pkg/pb/* $(OUT_DIR)/pb/; \
		rm -rf $(OUT_DIR)/example.com; \
	fi
	@echo "Running go mod tidy..."
	-$(GO_CMD) mod tidy || echo "Warning: go mod tidy had issues, but continuing..."

# Build the CLI tool
build-cli: proto-gen
	@echo "Building CLI tool..."
	$(GO_CMD) build -o bin/tinypenguin-cli ./cmd/tinypenguin-cli

# Build the gRPC server
build-server: proto-gen
	@echo "Building gRPC server..."
	$(GO_CMD) build -o bin/tinypenguin ./cmd/tinypenguin

# Build all components
build: build-cli build-server
	@echo "Build completed successfully!"

# Install the CLI tool globally
install: build-cli
	@echo "Installing CLI tool to /usr/local/bin..."
	sudo cp bin/tinypenguin-cli /usr/local/bin/tinypenguin-cli
	@echo "Installation completed!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf $(OUT_DIR)/pb/

# Test the code
test:
	@echo "Running tests..."
	$(GO_CMD) test ./...

# Check if protobuf compiler is installed
check-protoc:
	@command -v protoc >/dev/null 2>&1 || { echo "protoc not found. Please install protobuf compiler."; exit 1; }

# Install protobuf compiler (Linux)
install-protoc-linux:
	@echo "Installing protobuf compiler..."
	wget https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-linux-x86_64.zip
	unzip protoc-$(PROTOC_VERSION)-linux-x86_64.zip -d protoc-temp
	sudo mv protoc-temp/bin/protoc /usr/local/bin/
	sudo mv protoc-temp/include/* /usr/local/include/
	rm -rf protoc-temp protoc-$(PROTOC_VERSION)-linux-x86_64.zip

# Install protobuf compiler (macOS)
install-protoc-macos:
	@echo "Installing protobuf compiler..."
	brew install protobuf

# Run the CLI tool
run-cli: build-cli
	./bin/tinypenguin-cli $(args)

# Run the server
run-server: build-server
	./bin/tinypenguin $(args)

# Development target - build and run CLI with test query
dev:
	$(MAKE) run-cli args="run \"Test query for tinyllama\""

# Format Go code
fmt:
	@echo "Formatting Go code..."
	$(GO_CMD) fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run

# Generate documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060

# Docker build target
docker-build:
	@echo "Building Docker image..."
	docker build -t tinypenguin:latest .

# Docker run target
docker-run: docker-build
	@echo "Running Docker container..."
	docker run -it --rm -p 50051:50051 tinypenguin:latest
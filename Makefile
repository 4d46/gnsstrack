# Project variables
BINARY_NAME=gnsstrack
MODULE_NAME=4d46.uk/gnsstrack
DIST_DIR=bin

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build build-linux-arm64 clean test help

all: build-linux-arm64

## build: Builds the binary for the current architecture (local development)
build:
	@echo "Building for local architecture..."
	@go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME) main.go

## build-linux-arm64: Cross-compiles the binary for Raspberry Pi (CM4 64-bit)
build-linux-arm64:
	@echo "Building for Linux ARM64 (Raspberry Pi)..."
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 main.go

## test: Runs all unit tests
test:
	@go test -v ./...

## clean: Removes build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf $(DIST_DIR)
	@rm -f $(BINARY_NAME)

## help: Displays this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

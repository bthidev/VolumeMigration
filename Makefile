.PHONY: build build-linux build-all install test test-coverage lint vet clean help

# Version information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)
BUILD_FLAGS := -ldflags "$(LDFLAGS)"

# Build targets
build:
	@echo "Building volume-migrator $(VERSION)..."
	go build $(BUILD_FLAGS) -o bin/volume-migrator ./cmd/volume-migrator
	@echo "Build complete: bin/volume-migrator"

build-linux:
	@echo "Building for Linux AMD64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o bin/volume-migrator-linux-amd64 ./cmd/volume-migrator

build-all:
	@echo "Building for all platforms..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o bin/volume-migrator-linux-amd64 ./cmd/volume-migrator
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o bin/volume-migrator-linux-arm64 ./cmd/volume-migrator
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o bin/volume-migrator-darwin-amd64 ./cmd/volume-migrator
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o bin/volume-migrator-darwin-arm64 ./cmd/volume-migrator
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o bin/volume-migrator-windows-amd64.exe ./cmd/volume-migrator
	@echo "All builds complete"

install:
	@echo "Installing volume-migrator..."
	go install $(BUILD_FLAGS) ./cmd/volume-migrator

# Test targets
test:
	@echo "Running tests..."
	go test ./... -v

test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out
	@echo ""
	@echo "To view HTML coverage report, run: go tool cover -html=coverage.out"

# Code quality targets
lint:
	@echo "Running golangci-lint..."
	golangci-lint run --timeout=5m

vet:
	@echo "Running go vet..."
	go vet ./...

# Utility targets
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "Clean complete"

version:
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Date: $(DATE)"

help:
	@echo "Volume Migrator - Makefile targets"
	@echo ""
	@echo "Build targets:"
	@echo "  make build       - Build for current platform"
	@echo "  make build-linux - Build for Linux AMD64"
	@echo "  make build-all   - Build for all platforms"
	@echo "  make install     - Install to GOPATH/bin"
	@echo ""
	@echo "Test targets:"
	@echo "  make test          - Run tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo ""
	@echo "Code quality:"
	@echo "  make lint - Run golangci-lint"
	@echo "  make vet  - Run go vet"
	@echo ""
	@echo "Utility:"
	@echo "  make clean   - Remove build artifacts"
	@echo "  make version - Show version information"
	@echo "  make help    - Show this help message"

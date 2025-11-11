# Makefile for code_agent
# A CLI coding assistant powered by Google ADK Go

# Variables
BINARY_NAME=code-agent
VERSION=1.0.0
BUILD_DIR=../bin
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Detect OS for platform-specific commands
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    OPEN_CMD=open
else
    OPEN_CMD=xdg-open
endif

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: all build clean test install uninstall run fmt lint vet deps help

# Default target
all: clean deps fmt vet build

## help: Display this help message
help:
	@echo "$(GREEN)code_agent Makefile$(NC)"
	@echo ""
	@echo "Available targets:"
	@echo "  $(YELLOW)make build$(NC)      - Build the application"
	@echo "  $(YELLOW)make test$(NC)       - Run tests"
	@echo "  $(YELLOW)make run$(NC)        - Build and run the application"
	@echo "  $(YELLOW)make clean$(NC)      - Remove build artifacts"
	@echo "  $(YELLOW)make install$(NC)    - Install the binary to GOPATH/bin"
	@echo "  $(YELLOW)make uninstall$(NC)  - Remove the binary from GOPATH/bin"
	@echo "  $(YELLOW)make fmt$(NC)        - Format the code"
	@echo "  $(YELLOW)make lint$(NC)       - Run linters"
	@echo "  $(YELLOW)make vet$(NC)        - Run go vet"
	@echo "  $(YELLOW)make deps$(NC)       - Download dependencies"
	@echo "  $(YELLOW)make deps-update$(NC) - Update dependencies"
	@echo "  $(YELLOW)make coverage$(NC)   - Run tests with coverage"
	@echo "  $(YELLOW)make check$(NC)      - Run all checks (fmt, vet, lint, test)"
	@echo ""

## build: Build the application binary
build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## build-debug: Build with debug symbols
build-debug:
	@echo "$(GREEN)Building $(BINARY_NAME) with debug symbols...$(NC)"
	$(GO) build $(GOFLAGS) -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)✓ Debug build complete$(NC)"

## test: Run tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	$(GO) test -v ./...
	@echo "$(GREEN)✓ Tests complete$(NC)"

## test-short: Run short tests only
test-short:
	@echo "$(GREEN)Running short tests...$(NC)"
	$(GO) test -v -short ./...
	@echo "$(GREEN)✓ Short tests complete$(NC)"

## coverage: Run tests with coverage report
coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"
	@echo "$(YELLOW)Opening coverage report...$(NC)"
	@$(OPEN_CMD) coverage.html 2>/dev/null || echo "Please open coverage.html manually"

## bench: Run benchmarks
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	$(GO) test -bench=. -benchmem ./...

## clean: Remove build artifacts and temporary files
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f coverage.out coverage.html
	$(GO) clean
	@echo "$(GREEN)✓ Clean complete$(NC)"

## install: Install the binary to GOPATH/bin
install: build
	@echo "$(GREEN)Installing $(BINARY_NAME)...$(NC)"
	$(GO) install $(LDFLAGS) .
	@echo "$(GREEN)✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)$(NC)"

## uninstall: Remove the binary from GOPATH/bin
uninstall:
	@echo "$(YELLOW)Uninstalling $(BINARY_NAME)...$(NC)"
	rm -f $(shell go env GOPATH)/bin/$(BINARY_NAME)
	@echo "$(GREEN)✓ Uninstalled$(NC)"

## run: Build and run the application
run: build
	@echo "$(GREEN)Running $(BINARY_NAME)...$(NC)"
	$(BUILD_DIR)/$(BINARY_NAME)

## fmt: Format the code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GO) fmt ./...
	@echo "$(GREEN)✓ Format complete$(NC)"

## vet: Run go vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	$(GO) vet ./...
	@echo "$(GREEN)✓ Vet complete$(NC)"

## lint: Run golangci-lint (requires golangci-lint to be installed)
lint:
	@echo "$(GREEN)Running linters...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "$(GREEN)✓ Lint complete$(NC)"; \
	else \
		echo "$(YELLOW)⚠ golangci-lint not installed. Install with:$(NC)"; \
		echo "  brew install golangci-lint  # macOS"; \
		echo "  or go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## deps: Download dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GO) mod download
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

## deps-tidy: Tidy up go.mod and go.sum
deps-tidy:
	@echo "$(GREEN)Tidying dependencies...$(NC)"
	$(GO) mod tidy
	@echo "$(GREEN)✓ Dependencies tidied$(NC)"

## deps-update: Update dependencies
deps-update:
	@echo "$(GREEN)Updating dependencies...$(NC)"
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

## deps-verify: Verify dependencies
deps-verify:
	@echo "$(GREEN)Verifying dependencies...$(NC)"
	$(GO) mod verify
	@echo "$(GREEN)✓ Dependencies verified$(NC)"

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "$(GREEN)✓ All checks passed$(NC)"

## display-test: Build and run a quick display test
display-test: build
	@echo "$(GREEN)Testing display with markdown...$(NC)"
	@echo "## Test Display\n\nThis is a **test** with \`code\`" | ./$(BINARY_NAME) --output-format=rich || true

## info: Display build information
info:
	@echo "$(GREEN)Build Information:$(NC)"
	@echo "  Binary: $(BINARY_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Go Version: $(shell go version)"
	@echo "  GOOS: $(shell go env GOOS)"
	@echo "  GOARCH: $(shell go env GOARCH)"
	@echo "  Build Dir: $(BUILD_DIR)"

## todo: Show TODO items in code
todo:
	@echo "$(GREEN)TODO items:$(NC)"
	@grep -rn "TODO\|FIXME\|XXX" --include="*.go" . || echo "No TODO items found"

## size: Show binary size
size: build
	@echo "$(GREEN)Binary size:$(NC)"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME) | awk '{print "  " $$5 " - " $$9}'

## watch: Watch for changes and rebuild (requires entr)
watch:
	@if command -v entr >/dev/null 2>&1; then \
		echo "$(GREEN)Watching for changes...$(NC)"; \
		find . -name "*.go" | entr -r make build; \
	else \
		echo "$(RED)Error: entr not installed$(NC)"; \
		echo "Install with: brew install entr"; \
	fi

## docker-build: Build Docker image (if Dockerfile exists)
docker-build:
	@if [ -f Dockerfile ]; then \
		echo "$(GREEN)Building Docker image...$(NC)"; \
		docker build -t $(BINARY_NAME):$(VERSION) .; \
		echo "$(GREEN)✓ Docker image built$(NC)"; \
	else \
		echo "$(YELLOW)No Dockerfile found$(NC)"; \
	fi

## release: Create a release build (optimized, no debug info)
release:
	@echo "$(GREEN)Building release version...$(NC)"
	$(GO) build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)✓ Release build complete$(NC)"
	@make size

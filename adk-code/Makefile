# Makefile for adk-code
# A CLI coding assistant powered by Google ADK Go

# Variables
BINARY_NAME=adk-code
VERSION_FILE=.version
VERSION=$(shell cat $(VERSION_FILE) 2>/dev/null || echo "0.0.0")
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

.PHONY: help all build clean test install uninstall run fmt lint vet deps \
         check coverage bench info todo size watch docker-build release \
         build-debug test-short deps-tidy deps-update deps-verify display-test

# Default target - show help
.DEFAULT_GOAL := help

## help: Show this help message
help:
	@echo ""
	@echo "$(GREEN)╔════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(GREEN)║         adk-code - AI Coding Assistant CLI (ADK Go)           ║$(NC)"
	@echo "$(GREEN)╚════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(GREEN)📋 QUICK START$(NC)"
	@echo "  $(YELLOW)make run$(NC)         - Build and run the application"
	@echo "  $(YELLOW)make check$(NC)       - Run all checks before committing"
	@echo "  $(YELLOW)make help$(NC)        - Show this message (default)"
	@echo ""
	@echo "$(GREEN)🔨 BUILD & INSTALLATION$(NC)"
	@echo "  $(YELLOW)make build$(NC)       - Compile the binary to ../bin"
	@echo "  $(YELLOW)make build-debug$(NC) - Build with debug symbols for debugging"
	@echo "  $(YELLOW)make release$(NC)     - Optimized release build"
	@echo "  $(YELLOW)make install$(NC)     - Install binary to GOPATH/bin"
	@echo "  $(YELLOW)make uninstall$(NC)   - Remove binary from GOPATH/bin"
	@echo "  $(YELLOW)make clean$(NC)       - Remove all build artifacts"
	@echo ""
	@echo "$(GREEN)✅ TESTING & QUALITY$(NC)"
	@echo "  $(YELLOW)make test$(NC)        - Run all tests"
	@echo "  $(YELLOW)make test-short$(NC)  - Run short tests only"
	@echo "  $(YELLOW)make coverage$(NC)    - Generate coverage report (opens HTML)"
	@echo "  $(YELLOW)make check$(NC)       - Run fmt, vet, lint, and test"
	@echo "  $(YELLOW)make bench$(NC)       - Run benchmarks"
	@echo ""
	@echo "$(GREEN)📝 CODE QUALITY$(NC)"
	@echo "  $(YELLOW)make fmt$(NC)         - Format code with gofmt"
	@echo "  $(YELLOW)make vet$(NC)         - Run go vet for suspicious code"
	@echo "  $(YELLOW)make lint$(NC)        - Run golangci-lint (install if needed)"
	@echo ""
	@echo "$(GREEN)📦 DEPENDENCIES$(NC)"
	@echo "  $(YELLOW)make deps$(NC)        - Download all dependencies"
	@echo "  $(YELLOW)make deps-tidy$(NC)   - Tidy go.mod and go.sum"
	@echo "  $(YELLOW)make deps-update$(NC) - Update all dependencies"
	@echo "  $(YELLOW)make deps-verify$(NC) - Verify dependencies integrity"
	@echo ""
	@echo "$(GREEN)🛠️  UTILITIES$(NC)"
	@echo "  $(YELLOW)make info$(NC)          - Show build information"
	@echo "  $(YELLOW)make version$(NC)       - Display current build version"
	@echo "  $(YELLOW)make version-bump$(NC)  - Manually increment build number"
	@echo "  $(YELLOW)make version-set V=X$(NC) - Set version to X (e.g., make version-set V=1.0.0.5)"
	@echo "  $(YELLOW)make size$(NC)          - Show binary size"
	@echo "  $(YELLOW)make todo$(NC)          - Find TODO items in code"
	@echo "  $(YELLOW)make watch$(NC)         - Watch files and rebuild on change"
	@echo "  $(YELLOW)make display-test$(NC)  - Test display formatting"
	@echo "  $(YELLOW)make docker-build$(NC)  - Build Docker image (if Dockerfile exists)"
	@echo ""
	@echo "$(YELLOW)ℹ️  Tips:$(NC)"
	@echo "  • Run '$(YELLOW)make check$(NC)' before committing code"
	@echo "  • Use '$(YELLOW)make watch$(NC)' for continuous development"
	@echo "  • Check '$(YELLOW)make coverage$(NC)' to see test coverage"
	@echo ""

## build: Build the application binary (optimized)
build:
	@echo "$(GREEN)🔨 Building $(BINARY_NAME)...$(NC)" && \
	mkdir -p $(BUILD_DIR) && \
	NEW_VER=$$(bash scripts/version.sh bump) && \
	echo "$(GREEN)📦 Version: $$NEW_VER$(NC)" && \
	$(GO) build $(GOFLAGS) -ldflags "-X adk-code/internal/app.AppVersion=$$NEW_VER" -o $(BUILD_DIR)/$(BINARY_NAME) . && \
	echo "$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)" && \
	ls -lh $(BUILD_DIR)/$(BINARY_NAME) | awk '{print "  Size: " $$5}' 2>/dev/null || true

## build-debug: Build with debug symbols for debugging with dlv
build-debug:
	@echo "$(GREEN)🐛 Building $(BINARY_NAME) with debug symbols...$(NC)" && \
	NEW_VER=$$(bash scripts/version.sh bump) && \
	echo "$(GREEN)📦 Version: $$NEW_VER$(NC)" && \
	$(GO) build $(GOFLAGS) -gcflags="all=-N -l" -ldflags "-X adk-code/internal/app.AppVersion=$$NEW_VER" -o $(BUILD_DIR)/$(BINARY_NAME) . && \
	echo "$(GREEN)✓ Debug build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## test: Run all tests with verbose output
test:
	@echo "$(GREEN)🧪 Running tests...$(NC)"
	$(GO) test -v ./...
	@echo "$(GREEN)✓ Tests complete$(NC)"

## test-short: Run short tests only (fast)
test-short:
	@echo "$(GREEN)⚡ Running short tests...$(NC)"
	$(GO) test -v -short ./...
	@echo "$(GREEN)✓ Short tests complete$(NC)"

## coverage: Run tests with coverage report (opens HTML in browser)
coverage:
	@echo "$(GREEN)📊 Running tests with coverage...$(NC)"
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"
	@echo "$(YELLOW)🌐 Opening coverage report...$(NC)"
	@$(OPEN_CMD) coverage.html 2>/dev/null || echo "Please open coverage.html manually"

## bench: Run benchmarks to measure performance
bench:
	@echo "$(GREEN)📈 Running benchmarks...$(NC)"
	$(GO) test -bench=. -benchmem ./...

## clean: Remove all build artifacts and temporary files
clean:
	@echo "$(YELLOW)🗑️  Cleaning build artifacts...$(NC)"
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f coverage.out coverage.html
	$(GO) clean
	@echo "$(GREEN)✓ Clean complete$(NC)"

## install: Install the binary to GOPATH/bin
install: build
	@echo "$(GREEN)📥 Installing $(BINARY_NAME)...$(NC)"
	$(GO) install $(LDFLAGS) .
	@echo "$(GREEN)✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)$(NC)"

## uninstall: Remove the binary from GOPATH/bin
uninstall:
	@echo "$(YELLOW)📤 Uninstalling $(BINARY_NAME)...$(NC)"
	rm -f $(shell go env GOPATH)/bin/$(BINARY_NAME)
	@echo "$(GREEN)✓ Uninstalled$(NC)"

## run: Build and run the application interactively
run: build
	@echo "$(GREEN)▶️  Running $(BINARY_NAME)...$(NC)"
	$(BUILD_DIR)/$(BINARY_NAME)

## fmt: Format code with gofmt
fmt:
	@echo "$(GREEN)✨ Formatting code...$(NC)"
	$(GO) fmt ./...
	@echo "$(GREEN)✓ Format complete$(NC)"

## vet: Run go vet to check for suspicious code
vet:
	@echo "$(GREEN)🔍 Running go vet...$(NC)"
	$(GO) vet ./...
	@echo "$(GREEN)✓ Vet complete$(NC)"

## lint: Run golangci-lint for comprehensive linting
lint:
	@echo "$(GREEN)🎯 Running linters...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "$(GREEN)✓ Lint complete$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint not installed. Install with:$(NC)"; \
		echo "  brew install golangci-lint  # macOS"; \
		echo "  or go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## deps: Download all dependencies
deps:
	@echo "$(GREEN)📥 Downloading dependencies...$(NC)"
	$(GO) mod download
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

## deps-tidy: Tidy up go.mod and go.sum
deps-tidy:
	@echo "$(GREEN)🧹 Tidying dependencies...$(NC)"
	$(GO) mod tidy
	@echo "$(GREEN)✓ Dependencies tidied$(NC)"

## deps-update: Update all dependencies to latest versions
deps-update:
	@echo "$(GREEN)⬆️  Updating dependencies...$(NC)"
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

## deps-verify: Verify dependencies integrity
deps-verify:
	@echo "$(GREEN)✅ Verifying dependencies...$(NC)"
	$(GO) mod verify
	@echo "$(GREEN)✓ Dependencies verified$(NC)"

## check: Run all quality checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo ""
	@echo "$(GREEN)╔════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(GREEN)║                 ✓ All checks passed!                           ║$(NC)"
	@echo "$(GREEN)╚════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""

## display-test: Test display with markdown formatting
display-test: build
	@echo "$(GREEN)🎨 Testing display with markdown...$(NC)"
	@echo "## Test Display\n\nThis is a **test** with \`code\`" | ./$(BINARY_NAME) --output-format=rich || true

## info: Display build and environment information
info:
	@echo ""
	@echo "$(GREEN)📋 Build Information:$(NC)"
	@echo "  Binary Name:  $(BINARY_NAME)"
	@echo "  Version:      $(VERSION)"
	@echo "  Go Version:   $(shell go version)"
	@echo "  GOOS:         $(shell go env GOOS)"
	@echo "  GOARCH:       $(shell go env GOARCH)"
	@echo "  Build Dir:    $(BUILD_DIR)"
	@echo "  GOPATH:       $(shell go env GOPATH)"
	@echo ""

## todo: Find TODO, FIXME, XXX comments in code
todo:
	@echo "$(GREEN)📌 TODO items:$(NC)"
	@grep -rn "TODO\|FIXME\|XXX" --include="*.go" . || echo "  No TODO items found"
	@echo ""

## size: Display binary size information
size:
	@echo "$(GREEN)📦 Binary size:$(NC)"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME) 2>/dev/null | awk '{print "  " $$5 " - " $$9}' || echo "  Binary not yet built. Run 'make build' first."
	@echo ""

## watch: Watch files for changes and rebuild automatically (requires entr)
watch:
	@if command -v entr >/dev/null 2>&1; then \
		echo "$(GREEN)👀 Watching for changes...$(NC)"; \
		find . -name "*.go" | entr -r make build; \
	else \
		echo "$(RED)✗ Error: entr not installed$(NC)"; \
		echo "Install with: brew install entr"; \
	fi

## docker-build: Build Docker image (requires Dockerfile)
docker-build:
	@if [ -f Dockerfile ]; then \
		echo "$(GREEN)🐳 Building Docker image...$(NC)"; \
		docker build -t $(BINARY_NAME):$(VERSION) .; \
		echo "$(GREEN)✓ Docker image built: $(BINARY_NAME):$(VERSION)$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  No Dockerfile found$(NC)"; \
	fi

## version: Display current build version
version:
	@echo "$(GREEN)📌 Current version: $(shell cat $(VERSION_FILE))$(NC)"

## version-bump: Manually increment build number
version-bump:
	@echo "$(GREEN)📈 Bumping version...$(NC)"
	@bash scripts/version.sh bump

## version-set: Set version to a specific number (usage: make version-set V=1.0.0.5)
version-set:
	@if [ -z "$(V)" ]; then \
		echo "$(RED)Error: Version required (usage: make version-set V=1.0.0.5)$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)📝 Setting version to $(V)...$(NC)"
	@bash scripts/version.sh set "$(V)"
	@echo "$(GREEN)✓ Version updated: $(shell cat $(VERSION_FILE))$(NC)"

## release: Create optimized release build (no debug info)
release:
	@echo "$(GREEN)🚀 Building release version...$(NC)" && \
	NEW_VER=$$(bash scripts/version.sh bump) && \
	echo "$(GREEN)📦 Version: $$NEW_VER$(NC)" && \
	$(GO) build -ldflags="-s -w -X adk-code/internal/app.AppVersion=$$NEW_VER" -o $(BUILD_DIR)/$(BINARY_NAME) . && \
	echo "$(GREEN)✓ Release build complete$(NC)" && \
	ls -lh $(BUILD_DIR)/$(BINARY_NAME) | awk '{print "  Size: " $$5}' 2>/dev/null || true && \
	echo "" && \
	echo "$(YELLOW)Ready for distribution!$(NC)" && \
	echo ""

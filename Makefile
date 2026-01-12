.PHONY: help build clean test test-coverage test-integration lint fmt vet install run docker-build docker-run release

# Build variables
BINARY_NAME=ainative-code
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w \
  -X github.com/AINative-studio/ainative-code/internal/cmd.version=$(VERSION) \
  -X github.com/AINative-studio/ainative-code/internal/cmd.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo \"none\") \
  -X github.com/AINative-studio/ainative-code/internal/cmd.buildDate=$(BUILD_DATE) \
  -X github.com/AINative-studio/ainative-code/internal/cmd.builtBy=makefile"

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Build tags for SQLite features
SQLITE_TAGS=sqlite_fts5

# Directories
BUILD_DIR=./build
CMD_DIR=./cmd/ainative-code

# Platform-specific builds
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

help: ## Display this help message
	@echo "$(COLOR_BOLD)AINative Code - Makefile Commands$(COLOR_RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "$(COLOR_BOLD)Usage:$(COLOR_RESET)\n  make $(COLOR_BLUE)<target>$(COLOR_RESET)\n\n$(COLOR_BOLD)Targets:$(COLOR_RESET)\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(COLOR_BLUE)%-20s$(COLOR_RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(COLOR_BOLD)%s$(COLOR_RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

build: ## Build the application
	@echo "$(COLOR_GREEN)Building $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -tags "$(SQLITE_TAGS)" $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "$(COLOR_GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

build-all: ## Build for all platforms
	@echo "$(COLOR_GREEN)Building for all platforms...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		OUTPUT_NAME=$(BUILD_DIR)/$(BINARY_NAME)-$$OS-$$ARCH; \
		if [ $$OS = "windows" ]; then OUTPUT_NAME=$$OUTPUT_NAME.exe; fi; \
		echo "Building for $$OS/$$ARCH..."; \
		GOOS=$$OS GOARCH=$$ARCH CGO_ENABLED=1 $(GOBUILD) -tags "$(SQLITE_TAGS)" $(LDFLAGS) -o $$OUTPUT_NAME $(CMD_DIR); \
	done
	@echo "$(COLOR_GREEN)All builds complete!$(COLOR_RESET)"

run: build ## Build and run the application
	@echo "$(COLOR_GREEN)Running $(BINARY_NAME)...$(COLOR_RESET)"
	$(BUILD_DIR)/$(BINARY_NAME)

install: ## Install the application to $GOPATH/bin
	@echo "$(COLOR_GREEN)Installing $(BINARY_NAME)...$(COLOR_RESET)"
	$(GOCMD) install -tags "$(SQLITE_TAGS)" $(LDFLAGS) $(CMD_DIR)
	@echo "$(COLOR_GREEN)Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)$(COLOR_RESET)"

clean: ## Remove build artifacts
	@echo "$(COLOR_YELLOW)Cleaning build artifacts...$(COLOR_RESET)"
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "$(COLOR_GREEN)Clean complete!$(COLOR_RESET)"

##@ Testing

test: ## Run unit tests
	@echo "$(COLOR_GREEN)Running tests...$(COLOR_RESET)"
	$(GOTEST) -tags "$(SQLITE_TAGS)" -v -race ./...

test-coverage: ## Run tests with coverage report
	@echo "$(COLOR_GREEN)Running tests with coverage...$(COLOR_RESET)"
	$(GOTEST) -tags "$(SQLITE_TAGS)" -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)Coverage report generated: coverage.html$(COLOR_RESET)"
	@$(GOCMD) tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

test-coverage-check: ## Run tests and verify 80% coverage threshold
	@echo "$(COLOR_GREEN)Running tests with coverage threshold check...$(COLOR_RESET)"
	@$(GOTEST) -tags "$(SQLITE_TAGS)" -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	THRESHOLD=80.0; \
	echo "Coverage: $${COVERAGE}%"; \
	echo "Threshold: $${THRESHOLD}%"; \
	if (( $$(echo "$$COVERAGE < $$THRESHOLD" | bc -l) )); then \
		echo "$(COLOR_RED)ERROR: Coverage $${COVERAGE}% is below threshold $${THRESHOLD}%$(COLOR_RESET)"; \
		exit 1; \
	fi; \
	echo "$(COLOR_GREEN)SUCCESS: Coverage $${COVERAGE}% meets threshold $${THRESHOLD}%$(COLOR_RESET)"

test-integration: ## Run integration tests
	@echo "$(COLOR_GREEN)Running integration tests...$(COLOR_RESET)"
	@if [ -d "tests/integration" ]; then \
		$(GOTEST) -tags "$(SQLITE_TAGS) integration" -v -timeout=10m ./tests/integration/...; \
	else \
		echo "$(COLOR_YELLOW)No integration tests directory found$(COLOR_RESET)"; \
	fi

test-integration-coverage: ## Run integration tests with coverage
	@echo "$(COLOR_GREEN)Running integration tests with coverage...$(COLOR_RESET)"
	@if [ -d "tests/integration" ]; then \
		$(GOTEST) -tags "$(SQLITE_TAGS) integration" -v -timeout=10m -coverprofile=integration-coverage.out -covermode=atomic ./tests/integration/...; \
		$(GOCMD) tool cover -func=integration-coverage.out | grep total | awk '{print "Integration test coverage: " $$3}'; \
	else \
		echo "$(COLOR_YELLOW)No integration tests directory found$(COLOR_RESET)"; \
	fi

test-benchmark: ## Run benchmark tests
	@echo "$(COLOR_GREEN)Running benchmark tests...$(COLOR_RESET)"
	$(GOTEST) -tags "$(SQLITE_TAGS)" -bench=. -benchmem ./...

test-e2e: build ## Run E2E tests
	@echo "$(COLOR_GREEN)Running E2E tests...$(COLOR_RESET)"
	@if [ -d "tests/e2e" ]; then \
		cd tests/e2e && $(GOTEST) -tags "$(SQLITE_TAGS)" -v -timeout=10m ./...; \
	else \
		echo "$(COLOR_YELLOW)No E2E tests directory found$(COLOR_RESET)"; \
	fi

test-e2e-short: build ## Run E2E tests in short mode (skips long tests)
	@echo "$(COLOR_GREEN)Running E2E tests (short mode)...$(COLOR_RESET)"
	@if [ -d "tests/e2e" ]; then \
		cd tests/e2e && $(GOTEST) -tags "$(SQLITE_TAGS)" -v -short -timeout=5m ./...; \
	else \
		echo "$(COLOR_YELLOW)No E2E tests directory found$(COLOR_RESET)"; \
	fi

test-e2e-verbose: build ## Run E2E tests with verbose output
	@echo "$(COLOR_GREEN)Running E2E tests (verbose)...$(COLOR_RESET)"
	@if [ -d "tests/e2e" ]; then \
		cd tests/e2e && $(GOTEST) -tags "$(SQLITE_TAGS)" -v -timeout=10m ./... 2>&1 | tee e2e-test-output.log; \
	else \
		echo "$(COLOR_YELLOW)No E2E tests directory found$(COLOR_RESET)"; \
	fi

test-e2e-clean: ## Clean E2E test artifacts
	@echo "$(COLOR_GREEN)Cleaning E2E test artifacts...$(COLOR_RESET)"
	@rm -rf tests/e2e/artifacts/*
	@rm -f tests/e2e/e2e-test-output.log
	@echo "$(COLOR_GREEN)E2E artifacts cleaned!$(COLOR_RESET)"

test-all: test test-integration test-e2e ## Run all tests (unit, integration, and E2E)
	@echo "$(COLOR_GREEN)All tests completed!$(COLOR_RESET)"

##@ Code Quality

lint: ## Run golangci-lint
	@echo "$(COLOR_GREEN)Running linter...$(COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "$(COLOR_YELLOW)golangci-lint not installed. Install with:$(COLOR_RESET)"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; \
		exit 1; \
	fi

fmt: ## Format Go code
	@echo "$(COLOR_GREEN)Formatting code...$(COLOR_RESET)"
	$(GOFMT) -s -w .
	@echo "$(COLOR_GREEN)Format complete!$(COLOR_RESET)"

fmt-check: ## Check if code is formatted
	@echo "$(COLOR_GREEN)Checking code formatting...$(COLOR_RESET)"
	@UNFORMATTED=$$($(GOFMT) -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "$(COLOR_YELLOW)The following files are not formatted:$(COLOR_RESET)"; \
		echo "$$UNFORMATTED"; \
		exit 1; \
	fi; \
	echo "$(COLOR_GREEN)All files are formatted!$(COLOR_RESET)"

vet: ## Run go vet
	@echo "$(COLOR_GREEN)Running go vet...$(COLOR_RESET)"
	$(GOVET) ./...

security: ## Run security checks with gosec
	@echo "$(COLOR_GREEN)Running security scan...$(COLOR_RESET)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -fmt=json -out=gosec-report.json ./...; \
		gosec ./...; \
	else \
		echo "$(COLOR_YELLOW)gosec not installed. Install with:$(COLOR_RESET)"; \
		echo "  go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
		exit 1; \
	fi

vuln-check: ## Check for vulnerabilities
	@echo "$(COLOR_GREEN)Checking for vulnerabilities...$(COLOR_RESET)"
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "$(COLOR_YELLOW)govulncheck not installed. Installing...$(COLOR_RESET)"; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
		govulncheck ./...; \
	fi

##@ Dependencies

deps: ## Download dependencies
	@echo "$(COLOR_GREEN)Downloading dependencies...$(COLOR_RESET)"
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "$(COLOR_GREEN)Dependencies updated!$(COLOR_RESET)"

deps-upgrade: ## Upgrade dependencies
	@echo "$(COLOR_GREEN)Upgrading dependencies...$(COLOR_RESET)"
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "$(COLOR_GREEN)Dependencies upgraded!$(COLOR_RESET)"

deps-verify: ## Verify dependencies
	@echo "$(COLOR_GREEN)Verifying dependencies...$(COLOR_RESET)"
	$(GOMOD) verify
	@echo "$(COLOR_GREEN)Dependencies verified!$(COLOR_RESET)"

##@ Docker

docker-build: ## Build Docker image
	@echo "$(COLOR_GREEN)Building Docker image...$(COLOR_RESET)"
	docker build -t ainative-code:$(VERSION) -t ainative-code:latest \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		.
	@echo "$(COLOR_GREEN)Docker image built!$(COLOR_RESET)"

docker-run: ## Run Docker container
	@echo "$(COLOR_GREEN)Running Docker container...$(COLOR_RESET)"
	docker run -it --rm ainative-code:latest

docker-push: docker-build ## Build and push Docker image to registry
	@echo "$(COLOR_GREEN)Pushing Docker image...$(COLOR_RESET)"
	docker tag ainative-code:$(VERSION) ghcr.io/ainative-studio/ainative-code:$(VERSION)
	docker tag ainative-code:latest ghcr.io/ainative-studio/ainative-code:latest
	docker push ghcr.io/ainative-studio/ainative-code:$(VERSION)
	docker push ghcr.io/ainative-studio/ainative-code:latest
	@echo "$(COLOR_GREEN)Docker image pushed!$(COLOR_RESET)"

##@ Release

release: clean build-all ## Create a release build
	@echo "$(COLOR_GREEN)Creating release $(VERSION)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)/release
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		BINARY=$(BUILD_DIR)/$(BINARY_NAME)-$$OS-$$ARCH; \
		if [ $$OS = "windows" ]; then BINARY=$$BINARY.exe; fi; \
		if [ $$OS != "windows" ]; then \
			tar -czf $(BUILD_DIR)/release/$(BINARY_NAME)-$$OS-$$ARCH.tar.gz -C $(BUILD_DIR) $$(basename $$BINARY); \
		else \
			cd $(BUILD_DIR) && zip release/$(BINARY_NAME)-$$OS-$$ARCH.zip $$(basename $$BINARY) && cd ..; \
		fi; \
		shasum -a 256 $$BINARY > $$BINARY.sha256; \
	done
	@echo "$(COLOR_GREEN)Release $(VERSION) created in $(BUILD_DIR)/release/$(COLOR_RESET)"

changelog: ## Generate changelog
	@echo "$(COLOR_GREEN)Generating changelog...$(COLOR_RESET)"
	@git log --pretty=format:"- %s (%h)" --no-merges $$(git describe --tags --abbrev=0 2>/dev/null || echo "")..HEAD

##@ CI/CD Simulation

ci: fmt-check vet lint test-coverage-check ## Run CI checks locally
	@echo "$(COLOR_GREEN)All CI checks passed!$(COLOR_RESET)"

pre-commit: fmt vet lint test ## Run pre-commit checks
	@echo "$(COLOR_GREEN)Pre-commit checks passed!$(COLOR_RESET)"

##@ Information

version: ## Display version information
	@echo "$(COLOR_BOLD)Version:$(COLOR_RESET) $(VERSION)"
	@echo "$(COLOR_BOLD)Build Date:$(COLOR_RESET) $(BUILD_DATE)"

info: ## Display build information
	@echo "$(COLOR_BOLD)AINative Code Build Information$(COLOR_RESET)"
	@echo "  Binary Name:  $(BINARY_NAME)"
	@echo "  Version:      $(VERSION)"
	@echo "  Build Date:   $(BUILD_DATE)"
	@echo "  Go Version:   $$(go version)"
	@echo "  Build Dir:    $(BUILD_DIR)"
	@echo "  Platforms:    $(PLATFORMS)"

.DEFAULT_GOAL := help

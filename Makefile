# uinputd-go Makefile
# Builds daemon, client (with embedded daemon), and provides installation commands

.PHONY: all build build-daemon build-client clean install install-daemon install-systemd uninstall test test-unit test-integration test-coverage test-bench help

# Build configuration
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

# Paths
DAEMON_BIN := bin/uinputd
CLIENT_BIN := bin/uinput-client
INSTALL_PREFIX ?= /usr/local
DAEMON_INSTALL_PATH := $(INSTALL_PREFIX)/bin/uinputd
CLIENT_INSTALL_PATH := $(INSTALL_PREFIX)/bin/uinput-client
SYSTEMD_SERVICE_PATH := /etc/systemd/system/uinputd.service
CONFIG_PATH := /etc/uinputd

# Colors for output (use printf for better shell compatibility)
BOLD := $(shell printf '\033[1m')
GREEN := $(shell printf '\033[32m')
BLUE := $(shell printf '\033[34m')
YELLOW := $(shell printf '\033[33m')
RESET := $(shell printf '\033[0m')

# Nerd Font symbols (matching internal/styles/styles.go)
ICON_CHECK := $(shell printf '\uf00c')
ICON_WARNING := $(shell printf '\uf071')

##@ General

all: build ## Build all binaries (daemon + client)

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "$(BOLD)Usage:$(RESET)\n  make $(BLUE)<target>$(RESET)\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(BLUE)%-20s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(BOLD)%s$(RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

build: build-daemon build-client ## Build both daemon and client

build-daemon: ## Build uinputd daemon
	@echo "$(BOLD)Building daemon...$(RESET)"
	@mkdir -p bin
	go build $(LDFLAGS) -o $(DAEMON_BIN) ./cmd/uinputd
	@echo "$(GREEN)$(ICON_CHECK)$(RESET) Daemon built: $(DAEMON_BIN)"

build-client: build-daemon ## Build uinput-client (embeds daemon binary)
	@echo "$(BOLD)Building client with embedded daemon...$(RESET)"
	@mkdir -p bin
	@mkdir -p cmd/uinput-client/embedded
	@# Copy daemon binary for embedding
	cp $(DAEMON_BIN) cmd/uinput-client/embedded/uinputd
	@# Copy config for embedding
	cp configs/uinputd.yaml cmd/uinput-client/embedded/uinputd.yaml
	@# Generate systemd service file for embedding
	@sed "s|@DAEMON_PATH@|/usr/local/bin/uinputd|g" systemd/uinputd.service.template > cmd/uinput-client/embedded/uinputd.service
	@# Build client with embedded files
	go build $(LDFLAGS) -o $(CLIENT_BIN) ./cmd/uinput-client
	@# Cleanup embedded directory
	rm -rf cmd/uinput-client/embedded
	@echo "$(GREEN)$(ICON_CHECK)$(RESET) Client built: $(CLIENT_BIN)"

clean: ## Remove built binaries and test artifacts
	@echo "$(BOLD)Cleaning build artifacts...$(RESET)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "$(GREEN)$(ICON_CHECK)$(RESET) Clean complete"

##@ Install

install: install-client ## Install client to /usr/local/bin

install-client: build-client ## Install uinput-client to /usr/local/bin
	@echo "$(BOLD)Installing uinput-client...$(RESET)"
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "$(YELLOW)$(ICON_WARNING)$(RESET)  Installing client requires root privileges"; \
		echo "Run: sudo make install-client"; \
		exit 1; \
	fi
	install -m 755 $(CLIENT_BIN) $(CLIENT_INSTALL_PATH)
	@echo "$(GREEN)$(ICON_CHECK)$(RESET) Client installed: $(CLIENT_INSTALL_PATH)"
	@echo ""
	@echo "$(BOLD)Next steps:$(RESET)"
	@echo "  1. Install daemon:          $(CLIENT_INSTALL_PATH) install daemon"
	@echo "  2. Install systemd service: $(CLIENT_INSTALL_PATH) install systemd-service"
	@echo "  3. Enable service:          sudo systemctl enable uinputd"
	@echo "  4. Start service:           sudo systemctl start uinputd"
	@echo "  5. Activate group:          newgrp input"

uninstall: ## Uninstall daemon, client, and systemd service
	@echo "$(BOLD)Uninstalling uinputd...$(RESET)"
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "$(YELLOW)$(ICON_WARNING)$(RESET)  Uninstalling requires root privileges"; \
		echo "Run: sudo make uninstall"; \
		exit 1; \
	fi
	@# Stop and disable service if running
	@if systemctl is-active --quiet uinputd 2>/dev/null; then \
		systemctl stop uinputd >/dev/null 2>&1; \
		echo "$(GREEN)$(ICON_CHECK)$(RESET) Service stopped"; \
	fi
	@if systemctl is-enabled --quiet uinputd 2>/dev/null; then \
		systemctl disable --quiet uinputd >/dev/null 2>&1; \
		echo "$(GREEN)$(ICON_CHECK)$(RESET) Service disabled"; \
	fi
	@# Remove socket file if it exists
	@if [ -e /run/uinputd.sock ]; then \
		rm -f /run/uinputd.sock; \
		echo "$(GREEN)$(ICON_CHECK)$(RESET) Socket removed: /run/uinputd.sock"; \
	fi
	@# Remove files
	@if [ -f $(DAEMON_INSTALL_PATH) ]; then \
		rm -f $(DAEMON_INSTALL_PATH); \
		echo "$(GREEN)$(ICON_CHECK)$(RESET) Daemon removed: $(DAEMON_INSTALL_PATH)"; \
	fi
	@if [ -f $(CLIENT_INSTALL_PATH) ]; then \
		rm -f $(CLIENT_INSTALL_PATH); \
		echo "$(GREEN)$(ICON_CHECK)$(RESET) Client removed: $(CLIENT_INSTALL_PATH)"; \
	fi
	@if [ -f $(SYSTEMD_SERVICE_PATH) ]; then \
		rm -f $(SYSTEMD_SERVICE_PATH); \
		echo "$(GREEN)$(ICON_CHECK)$(RESET) Service file removed: $(SYSTEMD_SERVICE_PATH)"; \
	fi
	@if systemctl daemon-reload 2>/dev/null; then \
		echo "$(GREEN)$(ICON_CHECK)$(RESET) Systemd reloaded"; \
	fi
	@echo "$(GREEN)$(ICON_CHECK)$(RESET) Uninstall complete"
	@echo "$(YELLOW)Note:$(RESET) Config directory $(CONFIG_PATH) left intact (remove manually if desired)"
	@echo "$(YELLOW)Note:$(RESET) The 'input' group is left intact (remove manually if desired: sudo groupdel input)"

##@ Testing

test: test-unit ## Run all tests

test-unit: ## Run unit tests
	@echo "$(BOLD)Running unit tests...$(RESET)"
	go test -v -race ./internal/... ./pkg/... ./cmd/...
	@echo "$(GREEN)$(ICON_CHECK)$(RESET) Unit tests passed"

test-integration: ## Run integration tests (requires Docker)
	@echo "$(BOLD)Running integration tests in Docker...$(RESET)"
	@if ! command -v docker &>/dev/null; then \
		echo "$(YELLOW)$(ICON_WARNING)$(RESET)  Docker not found, skipping integration tests"; \
		exit 1; \
	fi
	docker build -f test/integration/Dockerfile -t uinputd-test .
	docker run --privileged uinputd-test go test -v ./test/integration/...
	@echo "$(GREEN)$(ICON_CHECK)$(RESET) Integration tests passed"

test-coverage: ## Generate test coverage report
	@echo "$(BOLD)Generating coverage report...$(RESET)"
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@COVERAGE=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}'); \
	echo "$(GREEN)$(ICON_CHECK)$(RESET) Total coverage: $(BOLD)$$COVERAGE$(RESET)"; \
	echo "Open coverage.html in browser for detailed report"

test-check-coverage: test-coverage ## Check if coverage meets 90% threshold
	@echo "$(BOLD)Checking coverage threshold...$(RESET)"
	@COVERAGE=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$COVERAGE < 90" | bc -l 2>/dev/null || echo "0") -eq 1 ]; then \
		echo "$(YELLOW)$(ICON_WARNING)$(RESET) Coverage $$COVERAGE% is below 90% threshold"; \
		exit 1; \
	else \
		echo "$(GREEN)$(ICON_CHECK)$(RESET) Coverage $$COVERAGE% meets 90% threshold"; \
	fi

test-bench: ## Run benchmarks
	@echo "$(BOLD)Running benchmarks...$(RESET)"
	go test -bench=. -benchmem ./internal/layouts/
	go test -bench=. -benchmem ./internal/server/
	@echo "$(GREEN)$(ICON_CHECK)$(RESET) Benchmarks complete"

##@ Development

mocks: ## Generate mocks using mockery
	@echo "$(BOLD)Generating mocks...$(RESET)"
	@if ! command -v mockery &>/dev/null; then \
		echo "$(YELLOW)$(ICON_WARNING)$(RESET)  mockery not found, install: go install github.com/vektra/mockery/v2@latest"; \
		exit 1; \
	fi
	mockery
	@echo "$(GREEN)$(ICON_CHECK)$(RESET) Mocks generated"

fmt: ## Format code
	@echo "$(BOLD)Formatting code...$(RESET)"
	go fmt ./...
	@echo "$(GREEN)$(ICON_CHECK)$(RESET) Format complete"

lint: ## Run linters
	@echo "$(BOLD)Running linters...$(RESET)"
	@if ! command -v golangci-lint &>/dev/null; then \
		echo "$(YELLOW)$(ICON_WARNING)$(RESET)  golangci-lint not found, install: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi
	golangci-lint run ./...
	@echo "$(GREEN)$(ICON_CHECK)$(RESET) Lint complete"

vet: ## Run go vet
	@echo "$(BOLD)Running go vet...$(RESET)"
	go vet ./...
	@echo "$(GREEN)$(ICON_CHECK)$(RESET) Vet complete"

mod-tidy: ## Tidy go.mod
	@echo "$(BOLD)Tidying modules...$(RESET)"
	go mod tidy
	@echo "$(GREEN)$(ICON_CHECK)$(RESET) Modules tidy"

check: fmt vet lint test-unit ## Run all checks (format, vet, lint, test)

##@ Utilities

version: ## Print version information
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

show-paths: ## Show installation paths
	@echo "$(BOLD)Installation Paths:$(RESET)"
	@echo "  Daemon:  $(DAEMON_INSTALL_PATH)"
	@echo "  Client:  $(CLIENT_INSTALL_PATH)"
	@echo "  Service: $(SYSTEMD_SERVICE_PATH)"
	@echo "  Config:  $(CONFIG_PATH)/uinputd.yaml"

run-daemon: build-daemon ## Run daemon locally (requires root)
	@echo "$(BOLD)Running daemon...$(RESET)"
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "$(YELLOW)$(ICON_WARNING)$(RESET)  Running daemon requires root privileges"; \
		echo "Run: sudo make run-daemon"; \
		exit 1; \
	fi
	$(DAEMON_BIN) --config configs/uinputd.yaml

run-client: build-client ## Run client with example command
	@echo "$(BOLD)Running client...$(RESET)"
	$(CLIENT_BIN) ping || echo "$(YELLOW)$(ICON_WARNING)$(RESET)  Daemon not running? Start with: sudo make run-daemon"

##@ Docker

docker-build: ## Build Docker image
	@echo "$(BOLD)Building Docker image...$(RESET)"
	docker build -t uinputd:$(VERSION) .
	@echo "$(GREEN)$(ICON_CHECK)$(RESET) Docker image built: uinputd:$(VERSION)"

docker-run: docker-build ## Run daemon in Docker (privileged)
	@echo "$(BOLD)Running in Docker...$(RESET)"
	docker run --rm --privileged -v /tmp:/tmp uinputd:$(VERSION)

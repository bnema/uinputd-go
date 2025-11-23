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

# Colors for output
BOLD := \033[1m
GREEN := \033[32m
BLUE := \033[34m
YELLOW := \033[33m
RESET := \033[0m

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
	@echo "$(GREEN)✓$(RESET) Daemon built: $(DAEMON_BIN)"

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
	@echo "$(GREEN)✓$(RESET) Client built: $(CLIENT_BIN)"

clean: ## Remove built binaries and test artifacts
	@echo "$(BOLD)Cleaning build artifacts...$(RESET)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "$(GREEN)✓$(RESET) Clean complete"

##@ Install

install: install-daemon install-client ## Install daemon and client to system

install-daemon: build-daemon ## Install uinputd daemon to /usr/local/bin
	@echo "$(BOLD)Installing daemon...$(RESET)"
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "$(YELLOW)⚠$(RESET)  Installing daemon requires root privileges"; \
		echo "Run: sudo make install-daemon"; \
		exit 1; \
	fi
	install -m 755 $(DAEMON_BIN) $(DAEMON_INSTALL_PATH)
	@# Create config directory
	mkdir -p $(CONFIG_PATH)
	@# Install default config if it doesn't exist
	@if [ ! -f $(CONFIG_PATH)/uinputd.yaml ]; then \
		install -m 644 configs/uinputd.yaml $(CONFIG_PATH)/uinputd.yaml; \
		echo "$(GREEN)✓$(RESET) Config installed: $(CONFIG_PATH)/uinputd.yaml"; \
	else \
		echo "$(YELLOW)⚠$(RESET)  Config exists, skipping: $(CONFIG_PATH)/uinputd.yaml"; \
	fi
	@echo "$(GREEN)✓$(RESET) Daemon installed: $(DAEMON_INSTALL_PATH)"

install-client: build-client ## Install uinput-client to /usr/local/bin
	@echo "$(BOLD)Installing client...$(RESET)"
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "$(YELLOW)⚠$(RESET)  Installing client requires root privileges"; \
		echo "Run: sudo make install-client"; \
		exit 1; \
	fi
	install -m 755 $(CLIENT_BIN) $(CLIENT_INSTALL_PATH)
	@echo "$(GREEN)✓$(RESET) Client installed: $(CLIENT_INSTALL_PATH)"

install-systemd: install-daemon ## Install and enable systemd service
	@echo "$(BOLD)Installing systemd service...$(RESET)"
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "$(YELLOW)⚠$(RESET)  Installing systemd service requires root privileges"; \
		echo "Run: sudo make install-systemd"; \
		exit 1; \
	fi
	@# Generate systemd service file
	@sed "s|@DAEMON_PATH@|$(DAEMON_INSTALL_PATH)|g" systemd/uinputd.service.template > /tmp/uinputd.service
	install -m 644 /tmp/uinputd.service $(SYSTEMD_SERVICE_PATH)
	rm /tmp/uinputd.service
	@# Reload systemd and enable service
	systemctl daemon-reload
	@echo "$(GREEN)✓$(RESET) Systemd service installed: $(SYSTEMD_SERVICE_PATH)"
	@echo ""
	@echo "$(BOLD)To enable and start the service:$(RESET)"
	@echo "  sudo systemctl enable uinputd"
	@echo "  sudo systemctl start uinputd"
	@echo ""
	@echo "$(BOLD)To check status:$(RESET)"
	@echo "  sudo systemctl status uinputd"

uninstall: ## Uninstall daemon, client, and systemd service
	@echo "$(BOLD)Uninstalling uinputd...$(RESET)"
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "$(YELLOW)⚠$(RESET)  Uninstalling requires root privileges"; \
		echo "Run: sudo make uninstall"; \
		exit 1; \
	fi
	@# Stop and disable service if running
	@if systemctl is-active --quiet uinputd; then \
		systemctl stop uinputd; \
		echo "$(GREEN)✓$(RESET) Service stopped"; \
	fi
	@if systemctl is-enabled --quiet uinputd; then \
		systemctl disable uinputd; \
		echo "$(GREEN)✓$(RESET) Service disabled"; \
	fi
	@# Remove files
	rm -f $(DAEMON_INSTALL_PATH)
	rm -f $(CLIENT_INSTALL_PATH)
	rm -f $(SYSTEMD_SERVICE_PATH)
	@if systemctl daemon-reload 2>/dev/null; then \
		echo "$(GREEN)✓$(RESET) Systemd reloaded"; \
	fi
	@echo "$(GREEN)✓$(RESET) Uninstall complete"
	@echo "$(YELLOW)Note:$(RESET) Config directory $(CONFIG_PATH) left intact (remove manually if desired)"

##@ Testing

test: test-unit ## Run all tests

test-unit: ## Run unit tests
	@echo "$(BOLD)Running unit tests...$(RESET)"
	go test -v -race ./internal/... ./pkg/... ./cmd/...
	@echo "$(GREEN)✓$(RESET) Unit tests passed"

test-integration: ## Run integration tests (requires Docker)
	@echo "$(BOLD)Running integration tests in Docker...$(RESET)"
	@if ! command -v docker &>/dev/null; then \
		echo "$(YELLOW)⚠$(RESET)  Docker not found, skipping integration tests"; \
		exit 1; \
	fi
	docker build -f test/integration/Dockerfile -t uinputd-test .
	docker run --privileged uinputd-test go test -v ./test/integration/...
	@echo "$(GREEN)✓$(RESET) Integration tests passed"

test-coverage: ## Generate test coverage report
	@echo "$(BOLD)Generating coverage report...$(RESET)"
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@COVERAGE=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}'); \
	echo "$(GREEN)✓$(RESET) Total coverage: $(BOLD)$$COVERAGE$(RESET)"; \
	echo "Open coverage.html in browser for detailed report"

test-check-coverage: test-coverage ## Check if coverage meets 90% threshold
	@echo "$(BOLD)Checking coverage threshold...$(RESET)"
	@COVERAGE=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$COVERAGE < 90" | bc -l 2>/dev/null || echo "0") -eq 1 ]; then \
		echo "$(YELLOW)❌$(RESET) Coverage $$COVERAGE% is below 90% threshold"; \
		exit 1; \
	else \
		echo "$(GREEN)✅$(RESET) Coverage $$COVERAGE% meets 90% threshold"; \
	fi

test-bench: ## Run benchmarks
	@echo "$(BOLD)Running benchmarks...$(RESET)"
	go test -bench=. -benchmem ./internal/layouts/
	go test -bench=. -benchmem ./internal/server/
	@echo "$(GREEN)✓$(RESET) Benchmarks complete"

##@ Development

fmt: ## Format code
	@echo "$(BOLD)Formatting code...$(RESET)"
	go fmt ./...
	@echo "$(GREEN)✓$(RESET) Format complete"

lint: ## Run linters
	@echo "$(BOLD)Running linters...$(RESET)"
	@if ! command -v golangci-lint &>/dev/null; then \
		echo "$(YELLOW)⚠$(RESET)  golangci-lint not found, install: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi
	golangci-lint run ./...
	@echo "$(GREEN)✓$(RESET) Lint complete"

vet: ## Run go vet
	@echo "$(BOLD)Running go vet...$(RESET)"
	go vet ./...
	@echo "$(GREEN)✓$(RESET) Vet complete"

mod-tidy: ## Tidy go.mod
	@echo "$(BOLD)Tidying modules...$(RESET)"
	go mod tidy
	@echo "$(GREEN)✓$(RESET) Modules tidy"

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
		echo "$(YELLOW)⚠$(RESET)  Running daemon requires root privileges"; \
		echo "Run: sudo make run-daemon"; \
		exit 1; \
	fi
	$(DAEMON_BIN) --config configs/uinputd.yaml

run-client: build-client ## Run client with example command
	@echo "$(BOLD)Running client...$(RESET)"
	$(CLIENT_BIN) ping || echo "$(YELLOW)⚠$(RESET)  Daemon not running? Start with: sudo make run-daemon"

##@ Docker

docker-build: ## Build Docker image
	@echo "$(BOLD)Building Docker image...$(RESET)"
	docker build -t uinputd:$(VERSION) .
	@echo "$(GREEN)✓$(RESET) Docker image built: uinputd:$(VERSION)"

docker-run: docker-build ## Run daemon in Docker (privileged)
	@echo "$(BOLD)Running in Docker...$(RESET)"
	docker run --rm --privileged -v /tmp:/tmp uinputd:$(VERSION)

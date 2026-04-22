BINARY     := kmgr
BIN_DIR    := bin
MAIN       := .
MODULE     := github.com/UnPoilTefal/kmgr
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -s -w \
	-X $(MODULE)/cmd.Version=$(VERSION) \
	-X $(MODULE)/cmd.Commit=$(COMMIT) \
	-X $(MODULE)/cmd.BuildDate=$(BUILD_DATE)

GO         := GOTOOLCHAIN=local go
GOFILES    := $(shell find . -name '*.go' -not -path './vendor/*')

.DEFAULT_GOAL := help

# ---------------------------------------------------------------------------
# Help
# ---------------------------------------------------------------------------
.PHONY: help
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# ---------------------------------------------------------------------------
# Development
# ---------------------------------------------------------------------------
.PHONY: build
build: ## Compile binary to bin/kmgr
	@mkdir -p $(BIN_DIR)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) $(MAIN)
	@echo "→ $(BIN_DIR)/$(BINARY)"

.PHONY: run
run: ## Run kmgr directly via go run (args: make run ARGS="import --help")
	$(GO) run $(MAIN) $(ARGS)

.PHONY: install
install: ## Install kmgr to GOPATH/bin (available in PATH)
	$(GO) install -ldflags "$(LDFLAGS)" $(MAIN)

# ---------------------------------------------------------------------------
# Quality
# ---------------------------------------------------------------------------
.PHONY: test
test: ## Run unit tests
	$(GO) test ./...

.PHONY: test-v
test-v: ## Run tests with verbose output
	$(GO) test -v ./...

.PHONY: test-race
test-race: ## Run tests with race condition detection
	$(GO) test -race ./...

.PHONY: vet
vet: ## Run go vet on all packages
	$(GO) vet ./...

.PHONY: lint
lint: ## Run golangci-lint (must be installed)
	@command -v golangci-lint &>/dev/null || { echo "golangci-lint not found — https://golangci-lint.run/usage/install/"; exit 1; }
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code (gofmt -w)
	@gofmt -w $(GOFILES)

.PHONY: check
check: vet test ## Run vet + tests (quick pre-commit check)

# ---------------------------------------------------------------------------
# Dependencies
# ---------------------------------------------------------------------------
.PHONY: tidy
tidy: ## Update go.mod and go.sum
	$(GO) mod tidy

.PHONY: deps
deps: ## Download dependencies
	$(GO) mod download

# ---------------------------------------------------------------------------
# Cross-compilation
# ---------------------------------------------------------------------------
.PHONY: build-all
build-all: ## Build for Linux, macOS (amd64/arm64), and Windows (amd64)
	@mkdir -p $(BIN_DIR)
	GOOS=linux   GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY)-linux-amd64   $(MAIN)
	GOOS=linux   GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY)-linux-arm64   $(MAIN)
	GOOS=darwin  GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY)-darwin-amd64  $(MAIN)
	GOOS=darwin  GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY)-darwin-arm64  $(MAIN)
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY)-windows-amd64.exe $(MAIN)
	@echo "→ Binaries in $(BIN_DIR)/"
	@ls -lh $(BIN_DIR)/

.PHONY: checksums
checksums: ## Generate SHA256 checksums for all binaries
	@cd $(BIN_DIR) && sha256sum * > checksums.txt && cat checksums.txt

.PHONY: release-build
release-build: clean build-all checksums ## Build all binaries and generate checksums for release

# ---------------------------------------------------------------------------
# Cleanup
# ---------------------------------------------------------------------------
.PHONY: clean
clean: ## Remove bin/ directory
	@rm -rf $(BIN_DIR)
	@echo "→ $(BIN_DIR)/ removed"

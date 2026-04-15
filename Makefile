APP_NAME   := gematria
SRC_DIR    := $(CURDIR)/src
BIN_DIR    := $(CURDIR)/bin
CMD_DIR    := ./cmd/gematria
LDFLAGS    := -s -w
HOST_OS    := $(shell go env GOOS)
HOST_ARCH  := $(shell go env GOARCH)
LOCAL_BIN  := $(BIN_DIR)/$(APP_NAME)-$(HOST_OS)-$(HOST_ARCH)

PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64 windows-arm64

.PHONY: help build build-all build-local \
        test test-race vet fmt fmt-check lint validate \
        clean clean-local run \
        $(addprefix build-,$(PLATFORMS))

help: ## List available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

build: validate clean-local build-local ## Build for current platform

build-all: validate clean $(addprefix build-,$(PLATFORMS)) ## Build for all platforms

build-local: | $(BIN_DIR) ## Build local binary
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=$(HOST_OS) GOARCH=$(HOST_ARCH) go build -ldflags="$(LDFLAGS)" -o $(LOCAL_BIN) $(CMD_DIR)

define build-target
build-$(1)-$(2): | $(BIN_DIR)
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME)-$(1)-$(2)$(3) $(CMD_DIR)
endef

$(eval $(call build-target,linux,amd64,))
$(eval $(call build-target,linux,arm64,))
$(eval $(call build-target,darwin,amd64,))
$(eval $(call build-target,darwin,arm64,))
$(eval $(call build-target,windows,amd64,.exe))
$(eval $(call build-target,windows,arm64,.exe))

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

run: ## Run locally with build flags
	cd $(SRC_DIR) && go run -ldflags="$(LDFLAGS)" $(CMD_DIR)

test: ## Run tests
	cd $(SRC_DIR) && go test ./...

test-race: ## Run tests with race detector
	cd $(SRC_DIR) && go test -race ./...

vet: ## Run go vet
	cd $(SRC_DIR) && go vet ./...

fmt: ## Format Go code
	cd $(SRC_DIR) && gofmt -w .

fmt-check: ## Check Go formatting
	@cd $(SRC_DIR) && test -z "$$(gofmt -l .)" || (echo "files not formatted:"; gofmt -l .; exit 1)

lint: ## Run linter
	cd $(SRC_DIR) && golangci-lint run ./...

validate: fmt-check vet lint test ## Run all checks

clean-local: ## Remove local binary
	rm -f $(LOCAL_BIN)

clean: ## Remove all build artifacts
	@[ -d "$(BIN_DIR)" ] && rm -f $(BIN_DIR)/$(APP_NAME)-* || true

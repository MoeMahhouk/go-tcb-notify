# ---- paths ----
BINDINGS_DIR := internal/registry/bindings
BINDINGS_FILE := $(BINDINGS_DIR)/registry.go

# Where your flashtestations repo lives (submodule or sibling)
FLASHTESTATION_DIR ?= ./flashtestations

# Path to the compiled artifact produced by `forge build`
REGISTRY_ARTIFACT ?= $(FLASHTESTATION_DIR)/foundry-out/FlashtestationRegistry.sol/FlashtestationRegistry.json

# Go package/type for the bindings
BINDINGS_PKG ?= bindings
BINDINGS_TYPE ?= FlashtestationRegistry

# ---- go binaries ----
BIN_INGEST   := bin/ingest-registry
BIN_PCS      := bin/fetch-pcs
BIN_EVAL     := bin/evaluate-quotes
BIN_VALIDATE := bin/validate-quotes

.PHONY: all bindings build tidy up down clean build-contracts check-tools schema

all: build

check-tools:
	@command -v forge >/dev/null 2>&1 || { echo "forge is not installed (foundry). See https://book.getfoundry.sh/"; exit 1; }
	@command -v abigen >/dev/null 2>&1 || { echo "abigen is not installed (go-ethereum). See https://geth.ethereum.org/docs/tools/abigen"; exit 1; }
	@command -v jq >/dev/null 2>&1 || { echo "jq is required"; exit 1; }

build-contracts: check-tools
	@echo "==> Building Flashtestation contracts..."
	cd $(FLASHTESTATION_DIR) && forge build

## Generate Go bindings from the ABI (requires: abigen on PATH)
bindings: build-contracts
	@echo "==> Generating Go bindings from artifact:"
	@test -f "$(REGISTRY_ARTIFACT)" || { echo "Artifact not found at $(REGISTRY_ARTIFACT)"; exit 1; }
	@mkdir -p $(BINDINGS_DIR)
	# Extract ABI portion into a temp file
	@tmp_abi=$$(mktemp); \
		jq -r '.abi' "$(REGISTRY_ARTIFACT)" > $$tmp_abi; \
		abigen --abi $$tmp_abi --pkg $(BINDINGS_PKG) --type $(BINDINGS_TYPE) --out $(BINDINGS_FILE); \
		rm -f $$tmp_abi
	@echo "==> Wrote $(BINDINGS_FILE)"

## Build all executables
build:
	mkdir -p bin
	go build -o $(BIN_INGEST) ./cmd/ingest-registry
	go build -o $(BIN_PCS)    ./cmd/fetch-pcs
	go build -o $(BIN_EVAL)   ./cmd/evaluate-quotes
	go build -o $(BIN_VALIDATE) ./cmd/validate-quotes

tidy:
	go mod tidy

schema:
	@echo "==> Schema is created automatically at service boot (internal/storage/clickhouse/schema.sql)."

## Docker Compose up/down (builds image)
up:
	docker compose up --build -d

down:
	docker compose down

clean:
	rm -rf bin
	rm -rf $(BINDINGS_DIR)

GOCACHE := $(CURDIR)/.cache/go-build
GOLANGCI_LINT_CACHE := $(CURDIR)/.cache/golangci-lint

export GOCACHE
export GOLANGCI_LINT_CACHE

.PHONY: help
help: ## Lists the available commands. Add a comment with '##' to describe a command.
	@grep -E '^[a-zA-Z_-].+:.*?## .*$$' $(MAKEFILE_LIST)\
		| sort\
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the cli
	@echo "Building the CLI..."
	@go build -o bin/cli main.go

.PHONY: run
run: ## Run the cli
	@echo "Running the CLI..."
	@go run main.go

.PHONY: test
test: ## Run the tests
	@echo "Running the tests..."
	@go test ./... -v

.PHONY: setup-ci
setup-ci: ## Install dependencies for CI
	@echo "Setting up CI dependencies..."
	@mkdir -p $(GOCACHE) $(GOLANGCI_LINT_CACHE)
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "golangci-lint could not be found, installing..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0; \
	else \
		echo "golangci-lint is already installed"; \
	fi

.PHONY: fmt
fmt: setup-ci ## Format the code
	@echo "Formatting the code..."
	@gofmt -s -w .
	@golangci-lint run --fix

.PHONY: lint
lint: setup-ci ## Run the linter
	@echo "Running the linter..."
	@golangci-lint run 

.PHONY: update-snap-all
update-snap-all: ## Update all the snaps
	@echo "Updating the snaps..."
	UPDATE_SNAPS=true go test ./... -v

.PHONY: install
install: ## Install the CLI
	@echo "Installing the CLI..."
	@go install ./main.go

.PHONY: nix-build-source
nix-build-source: ## Build the source code using Nix
	@echo "Building the source code using Nix..."
	@nix build .#source

.PHONY: nix-build-nightly
nix-build-nightly: ## Build the nightly using Nix
	@echo "Building the nightly using Nix..."
	@nix build .#nightly

.PHONY: nix-build
nix-build: ## Build the cli using Nix
	@echo "Building the cli using Nix..."
	@nix build .#default

.PHONY: nix-build-all
nix-build-all: nix-build-source nix-build-nightly nix-build ## Build all using Nix
	@echo "Building all using Nix..."

.PHONY: git-hooks-pre-push
git-hooks-pre-push: ## Set up git hooks and run
	echo "Pre-push git hooks set up"
	bash scripts/git-hooks/pre-push

.PHONY: tail-log-truncate
tail-log-truncate: ## Tail the log file and truncate it when it exceeds 10MB
	@echo "Tailing the log file and truncating if it exceeds 10MB..."
	truncate -s 0 /tmp/aerospace-scratchpad.log && tail -f /tmp/aerospace-scratchpad.log

.PHONY: tail-log
tail-log: ## Tail the log file
	@echo "Tailing the log file..."
	tail -f /tmp/aerospace-scratchpad.log

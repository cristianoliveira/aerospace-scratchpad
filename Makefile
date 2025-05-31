.PHONY: help
help: ## Lists the available commands. Add a comment with '##' to describe a command.
	@grep -E '^[a-zA-Z_-].+:.*?## .*$$' $(MAKEFILE_LIST)\
		| sort\
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


.PHONY: run
run: ## Run the cli
	@echo "Running the CLI..."
	@go run main.go

.PHONY: test
test: ## Run the tests
	@echo "Running the tests..."
	@go test ./... -v

.PHONY: fmt
fmt: ## Format the code
	@echo "Formatting the code..."
	@gofmt -s -w .
	@golangci-lint run --fix

.PHONY: lint
lint: ## Run the linter
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

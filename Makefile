.PHONY: build install clean test run help

# Build variables
BINARY_NAME=muxyard
BUILD_DIR=bin
CMD_DIR=cmd/muxyard
VERSION?=dev
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
help: ## Show this help message
	@echo "Muxyard - Tmux Session Manager"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

install: build ## Install the binary to /usr/local/bin
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete!"

install-user: build ## Install the binary to ~/bin
	@echo "Installing $(BINARY_NAME) to ~/bin..."
	@mkdir -p ~/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) ~/bin/
	@echo "Installation complete! Make sure ~/bin is in your PATH."

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

test: ## Run tests
	go test -v ./...

run: build ## Build and run the application
	./$(BUILD_DIR)/$(BINARY_NAME)

deps: ## Download dependencies
	go mod download
	go mod tidy

fmt: ## Format code
	go fmt ./...

lint: ## Run linter (requires golangci-lint)
	golangci-lint run

check: fmt lint test ## Run all checks (format, lint, test)

dev: ## Development build with race detection
	go build -race $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-dev ./$(CMD_DIR)

.DEFAULT_GOAL := help
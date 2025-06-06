BIN_DIR := ./bin
CMD_DIR := ./
LOGS_DIR := ./logs
MAIN_FILE := ./main.go
VERSION := 1.1.0

ifeq ($(OS),Windows_NT)
	BINARY_NAME := cookieclient.exe
	RM_CMD := if exist
	RM_DIR_CMD := rmdir /s /q
	MKDIR_CMD := if not exist "$@" mkdir
	ECHO_CMD := echo
	PATHSEP := \\
else
	BINARY_NAME := cookieclient
	RM_CMD := rm -rf
	RM_DIR_CMD := rm -rf
	MKDIR_CMD := mkdir -p
	ECHO_CMD := echo -e
	PATHSEP := /
endif

# Color codes
RESET := \033[0m
BOLD := \033[1m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
CYAN := \033[36m

help:
	@$(ECHO_CMD) "$(BOLD)Available commands:$(RESET)"
	@$(ECHO_CMD) "  $(CYAN)make build$(RESET)     - Compile the Go project"
	@$(ECHO_CMD) "  $(CYAN)make run$(RESET)       - Compile and run the binary"
	@$(ECHO_CMD) "  $(CYAN)make test$(RESET)      - Run tests"
	@$(ECHO_CMD) "  $(CYAN)make install$(RESET)   - Install the binary locally"
	@$(ECHO_CMD) "  $(CYAN)make version$(RESET)   - Show the version"
	@$(ECHO_CMD) "  $(CYAN)make clean$(RESET)     - Remove binaries and logs"
	@$(ECHO_CMD) "  $(CYAN)make build-prod-linux$(RESET) - Compile the Go project for Linux production"
	@$(ECHO_CMD) "  $(CYAN)make build-prod-windows$(RESET) - Compile the Go project for Windows production"
	@$(ECHO_CMD) "  $(CYAN)make lint$(RESET)      - Lint the Go files"
	@$(ECHO_CMD) "  $(CYAN)make fmt$(RESET)       - Format the Go files"
	@$(ECHO_CMD) "  $(CYAN)make help$(RESET)      - Show this message"
	@$(ECHO_CMD) "  $(CYAN)make build-windows$(RESET) - Compile for Windows"
	@$(ECHO_CMD) "  $(CYAN)make build-linux$(RESET)   - Compile for Linux"
	@$(ECHO_CMD) "  $(CYAN)make run-prod$(RESET)   - Run the production build"

build:
	@$(ECHO_CMD) "$(CYAN)[*] Building $(BOLD)...$(RESET)"
	@$(MKDIR_CMD) $(BIN_DIR)
	@go build -o $(BIN_DIR)$(PATHSEP)$(BINARY_NAME) $(CMD_DIR)$(MAIN_FILE)
	@$(ECHO_CMD) "$(GREEN)[+] Build complete!$(RESET)"

build-windows:
	@$(ECHO_CMD) "$(CYAN)[*] Building $(BOLD)for Windows...$(RESET)"
	@$(MKDIR_CMD) $(BIN_DIR)
	@GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)$(PATHSEP)cookieclient.exe $(CMD_DIR)$(MAIN_FILE)
	@$(ECHO_CMD) "$(GREEN)[+] Windows build complete!$(RESET)"

build-linux:
	@$(ECHO_CMD) "$(CYAN)[*] Building $(BOLD)for Linux...$(RESET)"
	@$(MKDIR_CMD) $(BIN_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)$(PATHSEP)cookieclient $(CMD_DIR)$(MAIN_FILE)
	@$(ECHO_CMD) "$(GREEN)[+] Linux build complete!$(RESET)"

build-prod-windows:
	@$(ECHO_CMD) "$(CYAN)[*] Building $(BOLD)for Windows production...$(RESET)"
	@$(MKDIR_CMD) $(BIN_DIR)
	@GOOS=windows GOARCH=amd64 \
		go build \
		-trimpath \
		-ldflags="-s -w -X main.Version=$(VERSION)" \
		-o $(BIN_DIR)$(PATHSEP)cookieclient.exe $(CMD_DIR)$(MAIN_FILE)
	@$(ECHO_CMD) "$(GREEN)[+] Windows production build complete!$(RESET)"

build-prod-linux:
	@$(ECHO_CMD) "$(CYAN)[*] Building $(BOLD)for Linux production...$(RESET)"
	@$(MKDIR_CMD) $(BIN_DIR)
	@GOOS=linux GOARCH=amd64 \
		go build \
		-race \
		-trimpath \
		-ldflags="-s -w -X main.Version=$(VERSION)" \
		-o $(BIN_DIR)$(PATHSEP)cookieclient $(CMD_DIR)$(MAIN_FILE)
	@$(ECHO_CMD) "$(GREEN)[+] Linux production build complete!$(RESET)"

build-prod: build-prod-linux build-prod-windows
	@$(ECHO_CMD) "$(GREEN)[+] Production builds complete!$(RESET)"

test:
	@$(ECHO_CMD) "$(CYAN)[*] Running tests $(BOLD)...$(RESET)"
	@go test ./...
	@$(ECHO_CMD) "$(GREEN)[+] Tests complete!$(RESET)"

install: build
	@$(ECHO_CMD) "$(CYAN)[*] Installing $(BOLD)...$(RESET)"
	@go install .
	@$(ECHO_CMD) "$(GREEN)[+] Install complete!$(RESET)"

run: build
	@$(ECHO_CMD) "$(CYAN)[*] Running $(BOLD)...$(RESET)"
	@$(BIN_DIR)$(PATHSEP)$(BINARY_NAME) $(ARGS)

run-prod: build-prod
	@$(ECHO_CMD) "$(CYAN)[*] Running production build $(BOLD)...$(RESET)"
	@$(BIN_DIR)$(PATHSEP)$(BINARY_NAME) $(ARGS)

version:
	@$(ECHO_CMD) "$(GREEN)[+] Version $(BINARY_NAME) v$(VERSION) $(RESET)"

clean:
ifeq ($(OS),Windows_NT)
	@$(ECHO_CMD) "$(YELLOW)[*] Cleaning binaries and db...$(RESET)"
	@if exist $(BIN_DIR) $(RM_DIR_CMD) $(BIN_DIR)
	@$(ECHO_CMD) "$(GREEN)[+] Clean complete!$(RESET)"
	@$(ECHO_CMD) "$(YELLOW)[*] Cleaning logs...$(RESET)"
	@if exist $(LOGS_DIR) $(RM_DIR_CMD) $(LOGS_DIR)
	@$(ECHO_CMD) "$(GREEN)[+] Clean complete!$(RESET)"
else
	@$(ECHO_CMD) "$(YELLOW)[*] Cleaning binaries and db...$(RESET)"
	@if [ -d $(BIN_DIR) ]; then $(RM_CMD) $(BIN_DIR)/*; fi
	@$(ECHO_CMD) "$(GREEN)[+] Clean complete!$(RESET)"
	@$(ECHO_CMD) "$(YELLOW)[*] Cleaning logs...$(RESET)"
	@if [ -d $(LOGS_DIR) ]; then $(RM_CMD) $(LOGS_DIR)/*; fi
	@$(ECHO_CMD) "$(GREEN)[+] Clean complete!$(RESET)"
endif

lint:
	@$(ECHO_CMD) "$(CYAN)[*] Linting $(BOLD)...$(RESET)"
	@if ! golangci-lint run; then \
		$(ECHO_CMD) "$(RED)[!] Linting failed!$(RESET)"; \
		exit 1; \
	fi
	@$(ECHO_CMD) "$(GREEN)[+] Linting complete!$(RESET)"

fmt: ## run gofumpt (if present) or gofmt
ifeq ($(OS),Windows_NT)
	@$(ECHO_CMD) "$(CYAN)[*] Formatting $(BOLD)...$(RESET)"
	@if where gofumpt >nul 2>&1; then \
		if ! gofumpt -w -d .; then \
			$(ECHO_CMD) "$(RED)[!] Formatting with gofumpt failed!$(RESET)"; \
			exit 1; \
		fi; \
	else \
		if ! go list -f {{.Dir}} ./... | xargs gofmt -w -s -d; then \
			$(ECHO_CMD) "$(RED)[!] Formatting with gofmt failed!$(RESET)"; \
			exit 1; \
		fi; \
	fi
else
	@$(ECHO_CMD) "$(CYAN)[*] Formatting $(BOLD)...$(RESET)"
	@if command -v gofumpt > /dev/null; then \
		if ! gofumpt -w -d .; then \
			$(ECHO_CMD) "$(RED)[!] Formatting with gofumpt failed!$(RESET)"; \
			exit 1; \
		fi; \
	else \
		if ! go list -f {{.Dir}} ./... | xargs gofmt -w -s -d; then \
			$(ECHO_CMD) "$(RED)[!] Formatting with gofmt failed!$(RESET)"; \
			exit 1; \
		fi; \
	fi
endif
	@$(ECHO_CMD) "$(GREEN)[+] Formatting complete!$(RESET)"

.PHONY: help build build-windows build-linux build-prod-windows build-prod-linux build-prod test install run run-prod version clean lint fmt

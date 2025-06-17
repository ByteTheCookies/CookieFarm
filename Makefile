# === COMMON VARIABLES ===
VERSION := 1.2.0

RESET := \033[0m
BOLD := \033[1m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
CYAN := \033[36m

# === SERVER VARIABLES ===
SERVER_BIN_DIR := ./bin
SERVER_CMD_DIR := ./cmd/server/
SERVER_LOGS_DIR := ./logs
SERVER_MAIN_FILE := main.go
SERVER_BINARY_NAME := cks
GOOS ?= linux
GOARCH ?= amd64

# === CLIENT VARIABLES ===
CLIENT_BIN_DIR := ./bin
CLIENT_CMD_DIR := ./cmd/client/
CLIENT_LOGS_DIR := ./logs
CLIENT_MAIN_FILE := main.go

ifeq ($(OS),Windows_NT)
	CLIENT_BINARY_NAME := ckc.exe
	RM_CMD := if exist
	RM_DIR_CMD := rmdir /s /q
	MKDIR_CMD := if not exist "$@" mkdir
	ECHO_CMD := echo
	PATHSEP := \\
else
	CLIENT_BINARY_NAME := ckc
	RM_CMD := rm -rf
	RM_DIR_CMD := rm -rf
	MKDIR_CMD := mkdir -p
	ECHO_CMD := echo -e
	PATHSEP := /
endif

# === HELP ===

help:
	@echo -e "$(BOLD)Available commands:$(RESET)"
	@echo -e "  $(CYAN)make build$(RESET)                  - Build both server and client"
	@echo -e "  $(CYAN)make server-build$(RESET)           - Build the server"
	@echo -e "  $(CYAN)make server-run$(RESET)             - Run the server"
	@echo -e "  $(CYAN)make server-install$(RESET)         - Install the server"
	@echo -e "  $(CYAN)make server-clean$(RESET)           - Clean server build"
	@echo -e "  $(CYAN)make server-watch$(RESET)           - Watch server files (via air)"
	@echo -e "  $(CYAN)make server-build-plugins$(RESET)   - Build server plugins"
	@echo -e "  $(CYAN)make tailwindcss-build$(RESET)      - Build Tailwind CSS"
	@echo -e "  $(CYAN)make tailwindcss-watch$(RESET)      - Watch Tailwind CSS"
	@echo -e "  $(CYAN)make minify$(RESET)                 - Minify JS"
	@echo -e ""
	@echo -e "  $(CYAN)make client-build$(RESET)           - Build the client"
	@echo -e "  $(CYAN)make client-run$(RESET)             - Run the client"
	@echo -e "  $(CYAN)make client-install$(RESET)         - Install the client"
	@echo -e "  $(CYAN)make client-test$(RESET)            - Run client tests"
	@echo -e "  $(CYAN)make client-install$(RESET)         - Install the client"
	@echo -e "  $(CYAN)make client-clean$(RESET)           - Clean client build"
	@echo -e "  $(CYAN)make client-build-linux$(RESET)     - Build client for Linux"
	@echo -e "  $(CYAN)make client-build-windows$(RESET)   - Build client for Windows"
	@echo -e "  $(CYAN)make client-build-prod$(RESET)      - Build client for production"

	@echo -e "  $(CYAN)make lint$(RESET)            - Lint client code"
	@echo -e "  $(CYAN)make fmt$(RESET)             - Format client code"

# === COMMON TARGETS ===

build: server-build client-build
	@$(ECHO_CMD) "$(GREEN)[+] Build complete for both server and client!$(RESET)"

# === SERVER TARGETS ===

server-build:
	@echo -e "$(CYAN)[*] Building server...$(RESET)"
	@mkdir -p $(SERVER_BIN_DIR)
	@go build -race -gcflags='github.com/ByteTheCookies/cookieserver/...="-m"' -o $(SERVER_BIN_DIR)/$(SERVER_BINARY_NAME) $(SERVER_CMD_DIR)/$(SERVER_MAIN_FILE)
	@echo -e "$(GREEN)[+] Server build complete!$(RESET)"

server-build-prod:
	@echo -e "$(CYAN)[*] Building server for production...$(RESET)"
	@mkdir -p $(SERVER_BIN_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -trimpath -ldflags="-s -w" -o $(SERVER_BIN_DIR)/$(SERVER_BINARY_NAME) $(SERVER_CMD_DIR)/$(SERVER_MAIN_FILE)
	@echo -e "$(GREEN)[+] Production build complete!$(RESET)"

server-run: server-build server-build-plugins minify
	@$(SERVER_BIN_DIR)/$(SERVER_BINARY_NAME) -c config.yml -d

server-install: tailwindcss-build server-build-prod server-build-plugins-prod
	@go install .

server-clean:
	@rm -rf $(SERVER_BIN_DIR)/* $(SERVER_LOGS_DIR)/*

server-build-plugins:
	@for file in $(shell find ./pkg/protocols -name '*.go' ! -name 'protocols.go'); do \
		if grep -q '^package main' "$$file"; then \
			filename=$$(basename $$file); \
			pluginname=$${filename%.go}; \
			go build -race -gcflags="all=-m" -buildmode=plugin -o "./pkg/protocols/$$pluginname.so" "$$file"; \
		else \
			echo "Skipping $$file: not a main package"; \
		fi; \
	done

server-build-plugins-prod:
	@for file in $(shell find ./pkg/protocols -name '*.go' ! -name 'protocols.go'); do \
		if grep -q '^package main' "$$file"; then \
			filename=$$(basename $$file); \
			pluginname=$${filename%.go}; \
			GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -buildmode=plugin -ldflags="-s -w" -o "./pkg/protocols/$$pluginname.so" "$$file"; \
		else \
			echo "Skipping $$file: not a main package"; \
		fi; \
	done

server-watch:
	@if command -v air > /dev/null; then air; else go install github.com/air-verse/air@latest && air; fi


# === CLIENT TARGETS ===

client-build:
	@$(ECHO_CMD) "$(CYAN)[*] Building client...$(RESET)"
	@$(MKDIR_CMD) $(CLIENT_BIN_DIR)
	@go build -o $(CLIENT_BIN_DIR)$(PATHSEP)$(CLIENT_BINARY_NAME) $(CLIENT_CMD_DIR)$(CLIENT_MAIN_FILE)
	@$(ECHO_CMD) "$(GREEN)[+] Client build complete!$(RESET)"

client-build-windows:
	@$(ECHO_CMD) "$(CYAN)[*] Building client for Windows...$(RESET)"
	@$(MKDIR_CMD) $(CLIENT_BIN_DIR)
	@GOOS=windows GOARCH=amd64 go build -o $(CLIENT_BIN_DIR)$(PATHSEP)$(CLIENT_BINARY_NAME) $(CLIENT_CMD_DIR)$(CLIENT_MAIN_FILE)
	@$(ECHO_CMD) "$(GREEN)[+] Windows build complete!$(RESET)"

client-build-linux:
	@$(ECHO_CMD) "$(CYAN)[*] Building client for Linux...$(RESET)"
	@$(MKDIR_CMD) $(CLIENT_BIN_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(CLIENT_BIN_DIR)$(PATHSEP)$(CLIENT_BINARY_NAME) $(CLIENT_CMD_DIR)$(CLIENT_MAIN_FILE)
	@$(ECHO_CMD) "$(GREEN)[+] Linux build complete!$(RESET)"

client-build-prod:
	@$(ECHO_CMD) "$(CYAN)[*] Building client for production...$(RESET)"
	@$(MAKE) client-build-linux
	@$(MAKE) client-build-windows
	@$(ECHO_CMD) "$(GREEN)[+] Production build complete!$(RESET)"

client-run: client-build
	@$(CLIENT_BIN_DIR)$(PATHSEP)$(CLIENT_BINARY_NAME)

client-install: client-build
	@sudo cp $(CLIENT_BIN_DIR)$(PATHSEP)$(CLIENT_BINARY_NAME) /usr/local/bin/$(CLIENT_BINARY_NAME)

client-test:
	@go test ./...

# === SHARED TOOLS ===

tailwindcss-build:
	./tools/tailwindcss -c ./internal/server/tailwind.config.js -i ./internal/server/assets/css/global.css -o ./internal/server/public/css/output.css --minify

tailwindcss-watch:
	./tools/tailwindcss -c ./internal/server/tailwind.config.js -i ./internal/server/assets/css/global.css -o ./internal/server/public/css/output.css --watch

minify:
	@uglifyjs ./internal/server/assets/js/*.js -o ./internal/server/public/js/output.min.js -c -m

lint:
	@if ! golangci-lint run; then exit 1; fi

fmt:
	@if command -v gofumpt > /dev/null; then gofumpt -w -d .; else go list -f {{.Dir}} ./... | xargs gofmt -w -s -d; fi

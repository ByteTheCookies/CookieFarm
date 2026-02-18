#!/usr/bin/env just --justfile
# === COMMON VARIABLES ===

VERSION := "1.2.1"
RESET := "\\033[0m"

# === SERVER VARIABLES ===

SERVER_BIN_DIR := "./bin"
SERVER_CMD_DIR := "./server"
SERVER_LOGS_DIR := "./logs"
SERVER_MAIN_FILE := "main.go"
SERVER_BINARY_NAME := "cks"
GOOS := "linux"
GOARCH := "amd64"

# === CLIENT VARIABLES ===

CLIENT_BIN_DIR := "./bin"
CLIENT_CMD_DIR := "./client"
CLIENT_MAIN_FILE := "main.go"

# CROSS PLATFORM BUILD VARIABLES
PATHSEP := if os() == "windows" { "\\" } else { "/" }
MKDIR_CMD := if os() == "windows" { "mkdir" } else { "mkdir -p" }
ECHO_CMD := if os() == "windows" { "echo" } else { "echo -e" }
CLIENT_BINARY_NAME := if os() == "windows" { "ckc.exe" } else { "ckc" }

# Display help information
help:
    @just --list

# === COMMON TARGETS ===

# Build both server and client
[group('build')]
build: server-build client-build
    @{{ ECHO_CMD }} "{{ GREEN }}[+] Build complete for both server and client!{{ RESET }}"

# === SERVER TARGETS ===

# Build server for development
[group('build')]
server-build:
    @echo -e "{{ CYAN }}[*] Building server...{{ RESET }}"
    @mkdir -p {{ SERVER_BIN_DIR }}
    @go build -o {{ SERVER_BIN_DIR }}/{{ SERVER_BINARY_NAME }} {{ SERVER_CMD_DIR }}/{{ SERVER_MAIN_FILE }}
    @echo -e "{{ GREEN }}[+] Server build complete!{{ RESET }}"

# Build server for production
[group('build')]
server-build-prod:
    @echo -e "{{ CYAN }}[*] Building server for production...{{ RESET }}"
    @mkdir -p {{ SERVER_BIN_DIR }}
    @GOOS={{ GOOS }} GOARCH={{ GOARCH }} \
      go build -race -trimpath -gcflags="github.com/ByteTheCookies/CookieFarm/...=-m" -ldflags="-s -w" -o {{ SERVER_BIN_DIR }}/{{ SERVER_BINARY_NAME }} {{ SERVER_CMD_DIR }}/{{ SERVER_MAIN_FILE }}
    @echo -e "{{ GREEN }}[+] Production build complete!{{ RESET }}"

# Run the server in development mode
[group('build')]
server-run: server-build server-build-plugins minify
    @{{ SERVER_BIN_DIR }}/{{ SERVER_BINARY_NAME }} -c -D

# Clean server binaries and logs
[group('dev')]
server-clean:
    @rm -rf {{ SERVER_BIN_DIR }}/* {{ SERVER_LOGS_DIR }}/*

# Build server plugins
[group('build')]
server-build-plugins:
    @for file in $(shell find ./pkg/protocols -name '*.go' ! -name 'protocols.go'); do \
     if grep -q '^package main' "$file"; then \
      filename=$(basename $$file); \
      pluginname=${filename%.go}; \
      go build -buildmode=plugin -o "./pkg/protocols/$pluginname.so" "$file"; \
     else \
      echo "Skipping $file: not a main package"; \
     fi; \
    done

# Build server plugins for production
[group('build')]
server-build-plugins-prod:
    @for file in $(shell find ./pkg/protocols -name '*.go' ! -name 'protocols.go'); do \
     if grep -q '^package main' "$file"; then \
      filename=$(basename $file); \
      pluginname=${filename%.go}; \
      GOOS={{ GOOS }} GOARCH={{ GOARCH }} go build -race -trimpath -gcflags="all=-m" -ldflags="-s -w" -buildmode=plugin -o "./pkg/protocols/$pluginname.so" "$file"; \
     else \
      echo "Skipping $file: not a main package"; \
     fi; \
    done

# Watch server files and rebuild on changes (requires air)
[group('dev')]
server-watch:
    @if command -v air > /dev/null; then air; else go install github.com/air-verse/air@latest && air; fi

# === CLIENT TARGETS ===

# Build client for development
[group('build')]
client-build:
    @{{ ECHO_CMD }} "{{ CYAN }}[*] Building client...{{ RESET }}"
    @{{ MKDIR_CMD }} {{ CLIENT_BIN_DIR }}
    @go build  -o {{ CLIENT_BIN_DIR }}{{ PATHSEP }}{{ CLIENT_BINARY_NAME }} {{ CLIENT_CMD_DIR }}/{{ CLIENT_MAIN_FILE }}
    @{{ ECHO_CMD }} "{{ GREEN }}[+] Client build complete!{{ RESET }}"

# Build client for Windows
[group('build')]
client-build-windows:
    @{{ ECHO_CMD }} "{{ CYAN }}[*] Building client for Windows...{{ RESET }}"
    @{{ MKDIR_CMD }} {{ CLIENT_BIN_DIR }}
    @GOOS=windows GOARCH=amd64 go build -o {{ CLIENT_BIN_DIR }}{{ PATHSEP }}{{ CLIENT_BINARY_NAME }} {{ CLIENT_CMD_DIR }}{{ CLIENT_MAIN_FILE }}
    @{{ ECHO_CMD }} "{{ GREEN }}[+] Windows build complete!{{ RESET }}"

# Build client for Linux
[group('build')]
client-build-linux:
    @{{ ECHO_CMD }} "{{ CYAN }}[*] Building client for Linux...{{ RESET }}"
    @{{ MKDIR_CMD }} {{ CLIENT_BIN_DIR }}
    @GOOS=linux GOARCH=amd64 go build -o {{ CLIENT_BIN_DIR }}{{ PATHSEP }}{{ CLIENT_BINARY_NAME }} {{ CLIENT_CMD_DIR }}{{ CLIENT_MAIN_FILE }}
    @{{ ECHO_CMD }} "{{ GREEN }}[+] Linux build complete!{{ RESET }}"

# Build client for production (Linux)
[group('build')]
client-build-linux-prod:
    @{{ ECHO_CMD }} "{{ CYAN }}[*] Building client for Linux production...{{ RESET }}"
    @{{ MKDIR_CMD }} {{ CLIENT_BIN_DIR }}
    @GOOS=linux GOARCH=amd64 go build -race -trimpath -gcflags="-m" -ldflags="-s -w" -o {{ CLIENT_BIN_DIR }}{{ PATHSEP }}{{ CLIENT_BINARY_NAME }} {{ CLIENT_CMD_DIR }}{{ CLIENT_MAIN_FILE }}
    @{{ ECHO_CMD }} "{{ GREEN }}[+] Linux production build complete!{{ RESET }}"

# Build client for production (Windows)
[group('build')]
client-build-windows-prod:
    @{{ ECHO_CMD }} "{{ CYAN }}[*] Building client for Windows production...{{ RESET }}"
    @{{ MKDIR_CMD }} {{ CLIENT_BIN_DIR }}
    @GOOS=windows GOARCH=amd64 go build -trimpath -gcflags="-m" -ldflags="-s -w" -o {{ CLIENT_BIN_DIR }}{{ PATHSEP }}{{ CLIENT_BINARY_NAME }} {{ CLIENT_CMD_DIR }}{{ CLIENT_MAIN_FILE }}
    @{{ ECHO_CMD }} "{{ GREEN }}[+] Windows production build complete!{{ RESET }}"

# Build client for production (all platforms)
[group('build')]
client-build-prod:
    @{{ ECHO_CMD }} "{{ CYAN }}[*] Building client for production...{{ RESET }}"
    @just client-build-linux-prod
    @just client-build-windows-prod
    @{{ ECHO_CMD }} "{{ GREEN }}[+] Production build complete!{{ RESET }}"

# Run the client
[group('dev')]
client-run: client-build
    @{{ CLIENT_BIN_DIR }}{{ PATHSEP }}{{ CLIENT_BINARY_NAME }}

# Install the client binary to /usr/local/bin and the virtual environment
[group('dev')]
client-install: client-build
    @sudo cp {{ CLIENT_BIN_DIR }}{{ PATHSEP }}{{ CLIENT_BINARY_NAME }} /usr/local/bin/{{ CLIENT_BINARY_NAME }}
    @sudo cp /usr/local/bin/{{ CLIENT_BINARY_NAME }} ~/.venv/bin/{{ CLIENT_BINARY_NAME }}

# Clean client binaries
[group('test')]
client-test:
    @go test ./...

# Start all the components for run mock tests mode for testing
[group('test')]
setup-tests num_containers="3" production_mode="false":
    cd ./scripts && ./setup.sh {{ num_containers }} {{ production_mode }}
# === SHARED TOOLS ===

# Build Tailwind CSS for production
[group('tools')]
tailwindcss-build:
    ./tools/tailwindcss -c ./server/tailwind.config.js -i ./server/assets/css/global.css -o ./server/public/css/output.css --minify

# Watch Tailwind CSS files and rebuild on changes
[group('tools')]
tailwindcss-watch:
    ./tools/tailwindcss -c ./server/tailwind.config.js -i ./server/assets/css/global.css -o ./server/public/css/output.css --watch

# Run the minify on the js files in the assets/js directory and output to public/js
[group('tools')]
minify:
    @uglifyjs ./server/assets/js/*.js -o ./server/public/js/output.min.js -c -m

# Lint the codebase using golangci-lint and apply fixes where possible
[group('tools')]
lint:
    @go work sync
    @go list -f \{\{.Dir\}\} -m | xargs golangci-lint run --fix

# Format the codebase using gofumpt
[group('tools')]
fmt:
    @go work sync
    @go list -f \{\{.Dir\}\} -m | xargs gofumpt -w -d

# Do a snapshot of the CPU, RAM, and Goroutines using pprof and open the web interface for each for the server
[group('dev')]
snapshot-cpu:
    @echo -e "{{ CYAN }}[*] Taking CPU snapshot...{{ RESET }}"
    go tool pprof -http=:6061 http://localhost:6060/debug/pprof/profile?seconds=30
    @echo -e "{{ GREEN }}[+] CPU snapshot complete!{{ RESET }}"

# Do a snapshot of the RAM using pprof and open the web interface for the server
[group('dev')]
snapshot-ram:
    @echo -e "{{ CYAN }}[*] Taking CPU snapshot...{{ RESET }}"
    go tool pprof -http=:6062 http://localhost:6060/debug/pprof/heap?seconds=30
    @echo -e "{{ GREEN }}[+] CPU snapshot complete!{{ RESET }}"

# Do a snapshot of the Goroutines using pprof and open the web interface for the server
[group('dev')]
snapshot-goroutine:
    @echo -e "{{ CYAN }}[*] Taking Goroutine snapshot...{{ RESET }}"
    go tool pprof -http=:6063 http://localhost:6060/debug/pprof/goroutine?debug=1
    @echo -e "{{ GREEN }}[+] Goroutine snapshot complete!{{ RESET }}"

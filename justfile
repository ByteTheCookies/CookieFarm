#!/usr/bin/env just --justfile

import 'env.just'
import 'cookiefarm/cookiefarm.just'
import 'exploiter/exploiter.just'
import 'cookiefarm/server/frontend/frontend.just'

# Display help information
help:
    @just --list

# === SHARED TOOLS ===

# Lint the codebase using golangci-lint and apply fixes where possible
[group('tools')]
[working-directory('cookiefarm')]
lint:
    @go work sync
    @go list -f \{\{.Dir\}\} -m | xargs golangci-lint run --fix

# Format the codebase using gofumpt
[group('tools')]
[working-directory('cookiefarm')]
fmt:
    @go work sync
    @go list -f \{\{.Dir\}\} -m | xargs gofumpt -w -d

# Format the codebase using gofumpt
[group('tools')]
[working-directory('cookiefarm/server')]
generate:
    @sqlc generate

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

[group('dev')]
[working-directory('cookiefarm')]
cookiefarm-release:
    @echo Creating branch for release...
    @./.github/release.sh release {{ VERSION }} || { echo -e "{{ RED }}[!] .github/release.sh failed. Aborting release.{{ RESET }}"; exit 1; }
    @goreleaser healthcheck || { echo -e "{{ RED }}[!] goreleaser healthcheck failed. Aborting release.{{ RESET }}"; exit 1; }
    @goreleaser release || { echo -e "{{ RED }}[!] goreleaser failed. Aborting release.{{ RESET }}"; exit 1; }
    @echo Release branch created and binaries built! Please review the release and publish it on GitHub.

[group('dev')]
exploiter-release:
    @echo Not implemented yet, but will build and upload the Python package to PyPI and Test PyPI

[group('dev')]
release:
    @just cookiefarm-release
    @just exploiter-release

[group('dev')]
[working-directory('cookiefarm')]
ghcr-push:
    @echo Building and pushing Docker image to GitHub Container Registry...
    @docker build -t ghcr.io/bytethecookies/cookiefarm:latest --build-arg VERSION=$(git describe --tags --abbrev=0) .
    @docker push ghcr.io/bytethecookies/cookiefarm:latest
    @echo Docker image pushed to GitHub Container Registry!

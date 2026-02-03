# cowardly — Brave Browser debloater for macOS
# Follows Standard Go Project Layout: https://github.com/golang-standards/project-layout

SHELL := /bin/bash
.DEFAULT_GOAL := help

# Build
BINARY  := cowardly
MAIN    := ./cmd/cowardly
BIN_DIR := bin
SCRIPTS := scripts
OUT     := $(BIN_DIR)/$(BINARY)

.PHONY: help build run test lint fmt prettier renovate clean install

help:
	@echo "cowardly — Brave Browser debloater for macOS"
	@echo ""
	@echo "Targets:"
	@echo "  build    build binary to $(OUT)"
	@echo "  run      build and run the TUI"
	@echo "  test     run tests"
	@echo "  lint     run golangci-lint"
	@echo "  fmt      format code and tidy modules"
	@echo "  prettier format markdown, json, yaml"
	@echo "  renovate run renovate (dry-run)"
	@echo "  clean    remove built binary"
	@echo "  install  build and install to $$(go env GOPATH)/bin"
	@echo ""
	@echo "Default: run 'make run' or 'make build' then '$(OUT)'"

build:
	@mkdir -p $(BIN_DIR)
	BINARY=$(OUT) MAIN=$(MAIN) $(SCRIPTS)/build.sh

run: build
	$(OUT)

test:
	go test ./...

lint:
	golangci-lint run

fmt:
	gofmt -s -w .
	go mod tidy

prettier:
	prettier --write "**/*.{md,json,yml,yaml}"

renovate:
	npx -y renovate --dry-run

clean:
	rm -f $(OUT)
	@rmdir $(BIN_DIR) 2>/dev/null || true

install: build
	go install $(MAIN)

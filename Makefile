# Cowardly — Brave Browser debloater for macOS
# Follows Standard Go Project Layout: https://github.com/golang-standards/project-layout

SHELL := /bin/zsh
.DEFAULT_GOAL := help

# Build
BINARY  := cowardly
MAIN    := ./cmd/cowardly
BIN_DIR := bin
SCRIPTS := scripts
OUT     := $(BIN_DIR)/$(BINARY)

.PHONY: help build run dev test lint lint-yaml fmt format format-check prettier renovate clean install bump-version

help:
	@echo "cowardly — Brave Browser debloater for macOS"
	@echo ""
	@echo "Targets:"
	@echo "  build         build binary to $(OUT)"
	@echo "  run           build and run the TUI"
	@echo "  dev           clean, then build and run the TUI"
	@echo "  test          run tests"
	@echo "  lint          run golangci-lint"
	@echo "  lint-yaml     run yamllint on YAML files"
	@echo "  fmt           format Go code and tidy modules"
	@echo "  format        format all files with Prettier (same version as CI)"
	@echo "  format-check  check formatting without making changes"
	@echo "  prettier      alias for format"
	@echo "  renovate      run renovate (dry-run)"
	@echo "  clean         remove built binary"
	@echo "  install       build and install to $$(go env GOPATH)/bin"
	@echo "  bump-version  create git tag and push (PART=patch|minor|major, default: patch)"
	@echo ""
	@echo "Default: run 'make run' or 'make build' then '$(OUT)'"

build:
	@mkdir -p $(BIN_DIR)
	BINARY=$(OUT) MAIN=$(MAIN) $(SCRIPTS)/build.sh

run: build
	$(OUT)

dev: clean run

# Run all package tests (*_test.go alongside code; no separate /test dir per Go convention)
test:
	go test ./...

lint:
	golangci-lint run

lint-yaml:
	@echo "Running yamllint..."
	@command -v yamllint >/dev/null 2>&1 || { echo "yamllint not found. Install with:  brew install yamllint  or  pip install yamllint"; exit 1; }
	@yamllint .
	@echo "yamllint done."

fmt:
	gofmt -s -w .
	go mod tidy

# Prettier version aligned with CI (.github/workflows/prettier.yml)
PRETTIER := npx --yes prettier@3.3.2

format:
	@echo "Formatting files with Prettier..."
	@$(PRETTIER) --write .

format-check:
	@echo "Checking file formatting..."
	@$(PRETTIER) --check .

prettier: format

renovate:
	npx -y renovate --dry-run

clean:
	rm -f $(OUT)
	@rmdir $(BIN_DIR) 2>/dev/null || true

install: build
	go install $(MAIN)

# Bump version: create annotated tag and push. Triggers release workflow.
# Usage: make bump-version [PART=patch|minor|major]
PART ?= patch
bump-version:
	@CURRENT=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	V=$${CURRENT#v}; V=$${V%%-*}; \
	MAJOR=$${V%%.*}; rest=$${V#*.}; \
	MINOR=$${rest%%.*}; rest=$${rest#*.}; \
	PATCH=$${rest%%.*}; \
	case "$(PART)" in \
	  major) MAJOR=$$((MAJOR+1)); MINOR=0; PATCH=0 ;; \
	  minor) MINOR=$$((MINOR+1)); PATCH=0 ;; \
	  patch) PATCH=$$((PATCH+1)) ;; \
	  *) echo "PART must be major, minor, or patch"; exit 1 ;; \
	esac; \
	NEW_TAG="v$${MAJOR}.$${MINOR}.$${PATCH}"; \
	echo "Creating tag $$NEW_TAG..."; \
	git tag -a "$$NEW_TAG" -m "Release $$NEW_TAG"; \
	git push origin "$$NEW_TAG"

#!/usr/bin/env bash
# Build the cowardly binary into BIN_DIR. Called from Makefile.
set -e
cd "$(dirname "$0")/.."
BINARY="${BINARY:-bin/cowardly}"
MAIN="${MAIN:-./cmd/cowardly}"
mkdir -p "$(dirname "$BINARY")"
# Embed version from git (e.g. v0.2.4 or v0.2.4-1-gabc1234-dirty), or "dev" if not in a repo / no tags
VERSION=$(git describe --tags --always --dirty 2>/dev/null) || VERSION="dev"
go build -ldflags "-X main.Version=${VERSION}" -o "$BINARY" "$MAIN"

#!/usr/bin/env bash
# Build the cowardly binary into BIN_DIR. Called from Makefile.
set -e
cd "$(dirname "$0")/.."
BINARY="${BINARY:-bin/cowardly}"
MAIN="${MAIN:-./cmd/cowardly}"
mkdir -p "$(dirname "$BINARY")"
go build -o "$BINARY" "$MAIN"

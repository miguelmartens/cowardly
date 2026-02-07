# Dev container for Cowardly — works on macOS (Docker), Linux, and WSL.
# Use for: build, test, lint, format. The app itself is macOS-only at runtime.
ARG GO_VERSION=1.25.6
FROM mcr.microsoft.com/devcontainers/base:bookworm

ARG GO_VERSION
ARG GOLANGCI_LINT_VERSION=v2.8.0

# Install system deps: make, zsh (for Makefile SHELL), yamllint, Node for Prettier
RUN apt-get update && apt-get install -y --no-install-recommends \
    make \
    zsh \
    ca-certificates \
    curl \
    git \
    yamllint \
    nodejs \
    npm \
    && rm -rf /var/lib/apt/lists/*

# Install Go (match go.mod: override GO_VERSION=1.25.6 when available)
RUN ARCH=$(dpkg --print-architecture) && \
    GOLANG_ARCH=${ARCH/amd64/linux-amd64} && \
    GOLANG_ARCH=${GOLANG_ARCH/arm64/linux-arm64} && \
    curl -sSL "https://go.dev/dl/go${GO_VERSION}.${GOLANG_ARCH}.tar.gz" -o /tmp/go.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm /tmp/go.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

# Install golangci-lint (v2 for .golangci.yml version "2")
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b /usr/local/bin "${GOLANGCI_LINT_VERSION}"

# Ensure go mod cache is writable and default to workspace
ENV GOMODCACHE=/go/pkg/mod
RUN mkdir -p /go/pkg/mod && chmod -R 777 /go

WORKDIR /workspace

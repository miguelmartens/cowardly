# Dev container for Cowardly — works on macOS (Docker), Linux, and WSL.
# Use for: build, test, lint, format. The app itself is macOS-only at runtime.
ARG GO_VERSION=1.25.7
FROM debian:bookworm-slim AS dev_containers_target_stage

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

# Install Go (satisfies go.mod 1.25.6+). Use case for POSIX sh (dash).
RUN ARCH=$(dpkg --print-architecture) && \
    case "$ARCH" in amd64) GOLANG_ARCH=linux-amd64;; arm64) GOLANG_ARCH=linux-arm64;; *) echo "Unsupported: $ARCH"; exit 1;; esac && \
    curl -sSL "https://go.dev/dl/go${GO_VERSION}.${GOLANG_ARCH}.tar.gz" -o /tmp/go.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm /tmp/go.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

# Install golangci-lint (v2 for .golangci.yml version "2")
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b /usr/local/bin "${GOLANGCI_LINT_VERSION}"

# Create vscode user so Cursor/VS Code can attach (devcontainer expects non-root)
RUN groupadd --gid 1000 vscode && useradd --uid 1000 --gid 1000 --shell /bin/zsh --create-home vscode
ENV GOMODCACHE=/go/pkg/mod
RUN mkdir -p /go/pkg/mod && chmod -R 777 /go
USER vscode
WORKDIR /workspace

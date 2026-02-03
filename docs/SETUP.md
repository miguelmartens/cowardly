# Setup Guide

Quick setup guide for the **cowardly** repository (Brave Browser debloater for macOS).

## Initial setup

### 1. Clone the repository

```bash
git clone https://github.com/your-username/cowardly.git
cd cowardly
```

### 2. Build and run

- **Requirements:** macOS, [Go 1.25.6+](https://go.dev/dl/), Brave Browser in `/Applications/Brave Browser.app`
- Build: `make build` → binary at `bin/cowardly`
- Run: `make run` or `./bin/cowardly`
- Install to `$(go env GOPATH)/bin`: `make install`

See the main [README](../README.md) for usage and presets.

## Enable GitHub automation

### Install Renovate

1. Go to [Renovate GitHub App](https://github.com/apps/renovate)
2. Click **Configure**
3. Select this repository (or your fork)
4. Grant the requested permissions
5. Click **Install**

Renovate will then:

- Open PRs for outdated **Go modules** (`go.mod` / `go.sum`)
- Open PRs for outdated **GitHub Actions** in `.github/workflows/`
- Run on the configured schedule (see below)
- Label PRs with `dependencies`

Configuration is in [renovate.json](../renovate.json) at the repo root (Go and gomod enabled, schedule: before 6am on Monday, etc.).

### Verify security scanning

**Gitleaks** is already configured via GitHub Actions:

- Runs on every push and pull request to `main`
- Fails the workflow if secrets are detected, so they cannot be merged

No extra setup is required. See [.github/workflows/README.md](../.github/workflows/README.md) for all workflows (gitleaks, prettier, yaml-lint).

## Maintenance

### Automatic updates (Renovate)

Renovate will:

- Create PRs when Go dependencies have updates
- Propose updates to GitHub Action versions in workflow files
- Run weekly (before 6am on Monday, per `renovate.json`)

Review and merge PRs as they appear to keep the project updated.

### Manual updates

- **Go modules:** `go get -u ./...` (or update specific modules), then `go mod tidy` and `make test`
- **Format / lint:** `make fmt`, `make format`, `make lint`, `make lint-yaml` before committing

## Security

### Secret management

- **Do not commit real secrets** — Use placeholders in committed files
- **Store secrets outside the repo** — No API keys or passwords in code or config
- **Rely on Gitleaks** — CI will fail if secrets are detected

### Before you push

Run `make test`, `make lint`, `make format-check`, and `make lint-yaml` so CI passes. See [Contributing](../README.md#contributing) in the README.

## Resources

- [Renovate documentation](https://docs.renovatebot.com/)
- [Renovate config (renovate.json)](../renovate.json)
- [Gitleaks](https://github.com/gitleaks/gitleaks)
- [GitHub Actions in this repo](../.github/workflows/README.md)

## Troubleshooting

### Renovate not creating PRs

- Confirm the [Renovate app](https://github.com/apps/renovate) is installed and has access to this repo
- Check the [Renovate dashboard](https://github.com/apps/renovate) for activity and errors
- Ensure [renovate.json](../renovate.json) is valid (e.g. run `make renovate` for a dry-run if you use it locally)

### Gitleaks failing

- Open the failed workflow run in the **Actions** tab
- Remove or replace any real secrets in the reported files; use placeholders only
- Re-run the workflow after fixing the commit

### Build or tests failing after a dependency update

- Run `go mod tidy` and `make test` locally
- If the new version breaks something, pin the previous version in `go.mod` or close the Renovate PR and open an issue

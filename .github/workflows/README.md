# GitHub Workflows

This directory contains GitHub Actions workflows for the **cowardly** repository (Brave Browser debloater for macOS, written in Go).

## Workflows

### Gitleaks

Scans the repository for secrets and sensitive information that may have been accidentally committed.

**Triggers:**
- Push to `main`
- Pull requests targeting `main`

**What it scans for:**
- API keys and tokens
- Passwords and credentials
- SSH keys
- And other common secret patterns

**Fail behavior:** The workflow fails if any secrets are detected, preventing them from being merged.

**Tip:** Never commit real secrets; use placeholders in committed files and store actual secrets outside the repo.

### Prettier

Checks that Markdown, JSON, YAML, and other supported files are formatted according to [Prettier](https://prettier.io/).

**Triggers:**
- Push to `main`
- Pull requests targeting `main`

**What it does:**
- Runs `prettier@3.3.2` via `npx --yes prettier@3.3.2 --check .`
- Fails if any file is not formatted (no auto-fix in CI)

**Local:** Run `make format-check` to match CI, or `make format` / `make prettier` to fix formatting.

### YAML Lint

Validates YAML syntax and style for YAML files in the repository using [yamllint](https://yamllint.readthedocs.io/).

**Configuration:** `.yamllint.yml` (at repository root). Some paths (e.g. `.github/workflows`, `renovate.json`) are ignored by that config.

**What it checks:**
- YAML syntax validity
- Indentation (2 spaces, indent-sequences)
- Line length (max 120, warning)
- Comments and empty lines
- Truthy values restricted to `true`/`false`/`on`/`off`

**Triggers:**
- Push to `main`
- Pull requests targeting `main`

**Local:** Run `make lint-yaml` to match CI.

### Release

Creates a GitHub Release with built binaries when you push a version tag.

**Triggers:**
- Push of a tag matching `v*` (e.g. `v1.0.0`, `v0.2.0`)

**What it does:**
- Builds the Go binary for **darwin/amd64** and **darwin/arm64** (macOS Intel and Apple Silicon)
- Creates a GitHub Release from the tag and attaches both binaries
- Generates release notes from the tag

**How to release:**
1. Bump version (e.g. in docs or go.mod if you track it there).
2. Commit, then create and push a tag: `git tag v1.0.0 && git push origin v1.0.0`
3. The workflow runs and publishes the release; download the binaries from the repository’s **Releases** page.

See **[docs/RELEASING.md](../../docs/RELEASING.md)** for full tagging and releasing instructions.

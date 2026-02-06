# Contributing to Cowardly

Thank you for your interest in contributing. This document explains how to get set up, run checks, and open a pull request.

## Requirements

- **macOS** (cowardly is macOS-only today)
- **Go 1.25.6+** — [Install Go](https://go.dev/dl/)
- Optional for local checks: **golangci-lint**, **prettier**, **yamllint** (CI will run these on your PR if you don’t have them)

## Getting started

1. **Fork** the repository on GitHub, then clone your fork:

   ```bash
   git clone https://github.com/YOUR_USERNAME/cowardly.git
   cd cowardly
   ```

2. **Build and run** to confirm everything works:

   ```bash
   make build
   make run
   ```

   See [README.md](README.md) for usage and [docs/SETUP.md](docs/SETUP.md) for repo setup (Renovate, Gitleaks, etc.).

## Making changes

1. **Create a branch** for your change:

   ```bash
   git checkout -b feat/my-feature
   ```

2. **Make your edits.** Some pointers:

   - **New presets** — Add a YAML file under [configs/presets/](configs/presets/). No Go code needed. See [docs/ADDING-PRESETS.md](docs/ADDING-PRESETS.md) for the format.
   - **Project layout** — [docs/PROJECT-LAYOUT.md](docs/PROJECT-LAYOUT.md) and [docs/FEATURES.md](docs/FEATURES.md) describe the codebase. Ideas for improvements are in [docs/FUTURE.md](docs/FUTURE.md).

3. **Run checks** before opening a PR:

   ```bash
   make test
   make lint
   make format-check
   make lint-yaml
   ```

   Fix any failures so CI passes.

4. **Commit and push** to your fork, then open a **pull request** with a short description of the change.

## Development commands

| Command             | Description                             |
| ------------------- | --------------------------------------- |
| `make build`        | Build binary to `bin/cowardly`          |
| `make run`          | Build and run the TUI                   |
| `make dev`          | Clean, then build and run               |
| `make test`         | Run tests                               |
| `make lint`         | Run golangci-lint                       |
| `make fmt`          | Format Go code and tidy modules         |
| `make format`       | Format Markdown/YAML/JSON with Prettier |
| `make format-check` | Check formatting (CI)                   |
| `make lint-yaml`    | Run yamllint on YAML files              |

## Pull request process

- Keep PRs focused; prefer smaller changes over large ones.
- For bugs or new features, opening an **issue** first is welcome (not required).
- CI runs on every push and PR: **gitleaks**, **prettier** (format check), **yaml-lint**. All must pass before merge.
- Maintainers will review and may request changes.

## Further reading

- [README.md](README.md) — Install, usage, presets
- [docs/ADDING-PRESETS.md](docs/ADDING-PRESETS.md) — How to add a preset
- [docs/PRIVACY-GUIDES.md](docs/PRIVACY-GUIDES.md) — Privacy Guides supplement
- [docs/PROJECT-LAYOUT.md](docs/PROJECT-LAYOUT.md) — Directory structure
- [docs/FUTURE.md](docs/FUTURE.md) — Possible improvements
- [.github/workflows/README.md](.github/workflows/README.md) — CI workflows

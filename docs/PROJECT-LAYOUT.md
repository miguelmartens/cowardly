# Project layout

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout).

## Directories in use

| Directory            | Purpose                                                                                                                                                                                      |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **cmd/cowardly**     | Main application entrypoint. Minimal `main` that imports from `internal` and runs the TUI or CLI.                                                                                            |
| **internal/**        | Private application code. Not importable by other projects.                                                                                                                                  |
| **internal/brave**   | Brave Browser preferences (macOS `defaults` read/write).                                                                                                                                     |
| **internal/config**  | Custom setting definitions for the TUI.                                                                                                                                                      |
| **internal/presets** | Loads preset definitions from embedded YAML in **configs/presets/** (one `.yaml` file per preset; add a file there and rebuild to add a preset). See [ADDING-PRESETS.md](ADDING-PRESETS.md). |
| **internal/ui**      | Bubble Tea TUI (model, update, view).                                                                                                                                                        |
| **configs/**         | Configuration templates. **configs/presets/** holds preset YAML files (embedded at build). See [configs/README.md](../configs/README.md).                                                    |
| **scripts/**         | Build and tool scripts; invoked by the root Makefile.                                                                                                                                        |
| **docs/**            | Design and user documentation (this file).                                                                                                                                                   |
| **assets/**          | Images and logos (e.g. `cowardly-logo.png`).                                                                                                                                                  |
| **bin/**             | Build output; executable from `make build`. Gitignored.                                                                                                                                      |

## Not used

- **pkg/** — No public, reusable library code.
- **api/** — No OpenAPI/specs.
- **web/** — No web assets.
- **build/** — No CI configs in-repo (e.g. GitHub Actions live under `.github/`).
- **vendor/** — Go modules only; no vendoring.
- **test/** — No external test apps or test data; unit tests live alongside code (`_test.go`).

## References

- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- [Organizing a Go module (go.dev)](https://go.dev/doc/modules/layout)

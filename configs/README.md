# configs

Configuration file templates and default configs.

## presets/

Brave debloat presets are **YAML** files here. Each file (e.g. `01-quick.yaml`) is one preset; order in the TUI is by filename. To add a preset, add a new `.yaml` file in **configs/presets/** and rebuild. See **[docs/ADDING-PRESETS.md](../docs/ADDING-PRESETS.md)** for the format and instructions.

This directory is reserved per the [Standard Go Project Layout](https://github.com/golang-standards/project-layout). Tool configs (e.g. `.golangci.yml`, `renovate.json`) remain at repository root by convention.

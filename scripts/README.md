# scripts

Scripts for build, install, and analysis. Used by the root Makefile to keep it small and simple.

- **build.sh** — build the cowardly binary (used by Makefile).
- **fresh-brave.sh** — completely remove Brave: quit app, delete all preferences/profile/caches, uninstall via Homebrew (`brew uninstall --cask brave-browser --zap`). Use before reinstalling for a clean state. Run `./scripts/fresh-brave.sh -h` for options (e.g. `-n` to skip Homebrew and only remove data + app).

See the [Standard Go Project Layout](https://github.com/golang-standards/project-layout) for details.

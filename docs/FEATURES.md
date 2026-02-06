# Implemented features

This document summarizes what Cowardly currently does. For possible future work, see [FUTURE.md](FUTURE.md).

## Core: policy application

- **Apply settings** — Write Brave policy keys (bool, integer, string) to macOS. Used by presets and Custom mode.
- **Managed preferences first** — Tries to write to `/Library/Managed Preferences/com.brave.Browser.plist` so Brave enforces policies (Rewards, Wallet, etc. hidden). Falls back to user preferences (`~/Library/Preferences/com.brave.Browser.plist`) if the user cancels the auth dialog or lacks admin rights.
- **Raw XML plist for managed** — Managed plist is generated as valid XML (not via `defaults write`) and copied into place with correct ownership and permissions.
- **Administrator privileges via AppleScript** — macOS authentication dialog (password or Touch ID) for writing to managed preferences; no password in the terminal.
- **Reset** — Removes all Brave policy settings: deletes user plist keys (and the plist file when empty) and removes the managed plist when present. Returns whether a managed plist existed and whether it was removed. Reset is blocked if Brave is running (user is told to quit Brave first).
- **Context timeouts** — `defaults` and `osascript` calls use timeouts (30s / 90s) to avoid hanging.

See [POLICY-ENFORCEMENT.md](POLICY-ENFORCEMENT.md) for the rationale and implementation details.

## TUI (Bubble Tea)

- **Apply a preset** — List of embedded presets (Quick Debloat, Maximum Privacy, Balanced, Performance, Developer, Strict Parental); choose one and apply.
- **Privacy Guides recommendations** — Apply [Privacy Guides](https://www.privacyguides.org/en/desktop-browsers/#brave) supplement on top of any preset (or Custom). TUI: choose base preset (Quick Debloat, Maximum Privacy, Custom if applied, etc.), then confirm. Contains only settings not in presets (Shields, P3A, De-AMP, etc.); no overlap. Config stored as `preset.<id>.settings` and `supplement.privacy_guides.settings` in `~/.config/cowardly/cowardly.yaml`.
- **Custom** — Toggle individual settings by category (Telemetry & Privacy, Privacy & Security, Brave Features, Performance & Bloat), then apply. Shortcuts: Space (toggle), Enter (apply), **a** (select all), **n** (select none).
- **View current settings** — Show which policy keys are set (user and managed when present).
- **Reset all to default** — Confirm with **y** / **Y** / Enter, then reset; clear messaging about managed vs user and Brave quit requirement.
- **Backups** — List backups, restore by path (Enter), or delete (d with confirmation).
- **Re-apply** — If the TUI detects that current settings differ from your saved desired state (e.g. reverted after restart), it shows a hint and you can press **R** to re-apply.
- **Exit** — Quit the TUI.
- **Brave orange styling** — Titles, active selections, and list components use Brave’s brand colors.
- **Message wrapping** — Long success/error messages (e.g. backup paths) wrap to terminal width.

## Desired state and re-apply

- **Config file** — When you apply a preset, Custom, or a file, Cowardly saves the applied state to `~/.config/cowardly/cowardly.yaml`. Format: `preset.<id>.settings` (presets or Custom), optionally `supplement.privacy_guides.settings` when Privacy Guides is applied. This is your "desired state."
- **Re-apply** — `--reapply` reads that config and re-applies the same settings. Use it after a restart when the organization or MDM has reverted your preferences.
- **Login hook** — `--install-login-hook` installs a Launch Agent (`~/Library/LaunchAgents/com.cowardly.reapply.plist`) that runs `cowardly --reapply` at every login, so your desired state is restored automatically.
- **TUI: reverted detection** — On startup, the TUI compares current Brave settings to the desired state. If they differ, it shows a message and lets you press **R** to re-apply without leaving the menu.

## CLI (non-interactive)

| Feature            | Flags                                                                                                                                                 |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| Apply preset       | `--apply`, `-a`, `--apply=<id>` (e.g. `max-privacy`, `balanced`)                                                                                      |
| Privacy Guides     | `--privacy-guides` (base from config or quick), `--privacy-guides=<base>` (e.g. max-privacy, custom)                                                  |
| Apply from file    | `--apply-file=<path>` (YAML with same `settings` format as presets)                                                                                   |
| Re-apply           | `--reapply` — re-apply last saved state from `~/.config/cowardly/`                                                                                    |
| Install login hook | `--install-login-hook` — run `--reapply` at every login                                                                                               |
| Dry run            | `--dry-run` (default: quick), `--dry-run=<id>`, `--dry-run=privacy-guides`, `--dry-run=privacy-guides:max-privacy`, `--dry-run=privacy-guides:custom` |
| Diff               | `--diff=<id>` — key-by-key difference (id can be `privacy-guides`, `privacy-guides:max-privacy`, or `privacy-guides:custom`)                          |
| Export             | `--export=<path>` — current settings to YAML                                                                                                          |
| Reset              | `--reset`, `-r`                                                                                                                                       |
| Current settings   | `--current`, `-c` — print current Brave policy settings                                                                                               |
| Backups            | `--backups`, `-b` (list), `--restore=<path>`, `--delete-backup=<path>`                                                                                |
| Help               | `--help`, `-h`                                                                                                                                        |

Apply and reset warn if Brave is running and block reset until Brave is quit.

## Presets

- **Six built-in presets** — Quick Debloat, Maximum Privacy, Balanced Privacy, Performance Focused, Developer, Strict Parental. Stored as YAML in `configs/presets/` and embedded at build time.
- **Supplements** — Stored in `configs/supplements/` (e.g. `supplements/privacy-guides/` for Privacy Guides). Apply on top of presets or Custom.
- **Preset format** — Each file: `id`, `name`, `description`, `settings` (list of `key`, `value`, `type`). Supported types: `bool`, `integer`, `string`. Preset keys validated with a simple name pattern.
- **Load errors** — Presets loaded with `AllWithError()`; load errors surface at startup.
- **Policy keys** — Support for telemetry, privacy, Brave features (Rewards, Wallet, VPN, AI, Tor, Sync), performance/bloat, proxy, startup, and extension allow/block lists (documented in [ADDING-PRESETS.md](ADDING-PRESETS.md)).

## Custom settings

- **Four categories** — Telemetry & Privacy, Privacy & Security, Brave Features, Performance & Bloat (order defined in `internal/config/settings.go`).
- **~28 toggleable settings** — Metrics, Safe Browsing, autofill, password manager, sign-in, WebRTC, QUIC, cookies, Do Not Track, SafeSearch, IPFS, incognito, Rewards, Wallet, VPN, AI Chat, Tor, Sync, background mode, recommendations, shopping list, PDF externally, translate, spellcheck, promotions, search suggestions, printing, default browser prompt, developer tools.
- **Same policy keys** as presets; values aligned with slimbrave-macos / bebrave.

## Backup and restore

- **Auto backup on apply/reset** — User plist copied to `~/Library/Application Support/cowardly/backups/<timestamp>-user.plist` before apply or reset.
- **List / restore / delete** — CLI flags and TUI Backups menu to list paths, restore from a backup, or delete a backup file.

## Brave detection

- **Brave version** — Read from app bundle via `defaults read` (shown by `--version` / `-v`).
- **Brave running** — Check if Brave process is running; used to warn before apply and to block reset until Brave is quit.

## Project and tooling

- **Layout** — [Standard Go Project Layout](https://github.com/golang-standards/project-layout): `cmd/cowardly`, `internal/brave`, `internal/config`, `internal/presets`, `internal/ui`, `configs/presets`, `configs/supplements`, `docs`, `scripts`, `assets`. See [PROJECT-LAYOUT.md](PROJECT-LAYOUT.md).
- **Makefile** — `build`, `run`, `test`, `lint` (golangci-lint v2), `fmt`, `format-check`, `prettier`, `lint-yaml`, `clean`, `install`.
- **CI (GitHub Actions)** — Gitleaks (secrets), Prettier (format check), yaml-lint (YAML syntax), release (macOS amd64/arm64 binaries on `v*` tags). See `.github/workflows/` and [RELEASING.md](RELEASING.md).
- **Scripts** — `scripts/fresh-brave.sh` for a full Brave wipe (uninstall, remove app and all Brave data); documented for cases where a reinstall does not clear old preferences.

## Documentation

- **README.md** — Install, usage (TUI + CLI), presets table, custom settings, project layout, development, contributing, disclaimer.
- **docs/INSTALL.md** — Install from a release (download, extract, run, install to PATH, upgrade).
- **docs/ADDING-PRESETS.md** — How to add presets (YAML format, keys reference, troubleshooting).
- **docs/PLATFORMS.md** — macOS-only today; possible Linux/Windows and policy mechanisms.
- **docs/POLICY-ENFORCEMENT.md** — Why managed preferences, implementation (raw plist, AppleScript), MDM note.
- **docs/PRIVACY-GUIDES.md** — Privacy Guides supplement: how it works, config format, TUI/CLI usage.
- **docs/PROJECT-LAYOUT.md** — Directory map and references.
- **docs/SETUP.md** — Repo setup, Renovate, Gitleaks, maintenance.
- **docs/RELEASING.md** — Tagging and release workflow.
- **docs/FUTURE.md** — Possible improvements and new features.
- **docs/FEATURES.md** — This file.

## Platform scope

- **macOS only** — Uses `defaults`, `~/Library/Preferences/com.brave.Browser.plist`, and `/Library/Managed Preferences/com.brave.Browser.plist`. No Linux or Windows support yet; see [PLATFORMS.md](PLATFORMS.md).

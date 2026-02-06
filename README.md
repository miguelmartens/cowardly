![](assets/cowardly-logo.png)

# Cowardly

**Cowardly** removes Brave’s bold features and returns it to a quiet, minimal, privacy-first browser on macOS. It is a small TUI (terminal UI) and CLI that applies Brave Browser policy preferences via macOS managed preferences (or `defaults`) so you can disable rewards, wallet, VPN, AI, telemetry, and other bloat without editing plists by hand.

Inspired by [SlimBrave](https://github.com/ltx0101/SlimBrave), [Debloat Brave Browser (macOS)](https://github.com/hi-one/hi-one/blob/main/Debloat-Brave-Browser-MacOS.md), [bebrave](https://github.com/ricardorodrigues-ca/bebrave), and [slimbrave-macos](https://github.com/vladandrei51/slimbrave-macos).

## Requirements

- **macOS** only today (uses `defaults` and `~/Library/Preferences/com.brave.Browser.plist`)
- **Go 1.25.6+** to build
- **Brave Browser** installed in `/Applications/Brave Browser.app` (tested with Brave stable; policy keys may vary by Brave version)

**Platform support:** Cowardly currently supports **macOS only**. Support for **Linux** and **Windows** may be added in the future; on those platforms Brave uses different policy mechanisms (e.g. JSON on Linux, registry/Group Policy on Windows). See **[docs/PLATFORMS.md](docs/PLATFORMS.md)** for details and contribution notes.

## Install

```bash
git clone https://github.com/miguelmartens/cowardly.git
cd cowardly
make build
./bin/cowardly
```

Or install into `$(go env GOPATH)/bin`:

```bash
make install
cowardly
```

To build and run from the repo (e.g. for development):

```bash
make run
```

For a clean build and run (removes the binary first):

```bash
make dev
```

For repository setup and Renovate automation, see **[docs/SETUP.md](docs/SETUP.md)**. For how we tag and publish releases, see **[docs/RELEASING.md](docs/RELEASING.md)**.

## Usage

### TUI (default)

Run with no arguments to start the interactive TUI:

```bash
cowardly
```

- **Apply a preset** — Choose a preset (Quick Debloat, Maximum Privacy, Balanced, Performance, Developer, Strict Parental) and apply it.
- **Custom** — Toggle individual settings by category (Telemetry, Privacy & Security, Brave Features, Performance & Bloat), then apply.
- **View current settings** — See which policy keys are set.
- **Reset all to default** — Remove all Brave policy settings (restore defaults).
- **Exit** — Quit.

After applying or resetting, **restart Brave Browser** for changes to take effect.

**Enforced policies:** Cowardly first tries to write to `/Library/Managed Preferences/com.brave.Browser.plist` so Brave enforces the policies (hides Rewards, Wallet, etc.). A **macOS authentication dialog** appears (password or Touch ID)—use that to approve; you don’t type the password in the terminal. If you cancel or don’t have admin rights, settings are written to user preferences only; Brave may still show those features. Reset may also show the dialog if the managed plist exists. See **[docs/POLICY-ENFORCEMENT.md](docs/POLICY-ENFORCEMENT.md)** for why this is needed and how it is implemented.

### CLI (non-interactive)

- **Apply a preset** — Quick Debloat (default) or by ID (e.g. `max-privacy`, `balanced`):

  ```bash
  cowardly --apply
  cowardly -a
  cowardly --apply=max-privacy
  ```

- **Apply from a YAML file** (same format as preset `settings`):

  ```bash
  cowardly --apply-file=./my-settings.yaml
  ```

- **Re-apply saved state** (e.g. after a restart when MDM or the organization has reverted settings). Cowardly saves your last-applied preset or file to `~/.config/cowardly/cowardly.yaml`; use `--reapply` to restore it:

  ```bash
  cowardly --reapply
  ```

- **Install login hook** — Run `cowardly --reapply` automatically at every login (installs a Launch Agent). Useful when your Mac is managed and policies are re-applied on boot:

  ```bash
  cowardly --install-login-hook
  ```

  To remove: `rm ~/Library/LaunchAgents/com.cowardly.reapply.plist`

- **Dry run / diff** — See what would be applied, or which keys would change:

  ```bash
  cowardly --dry-run
  cowardly --dry-run=balanced
  cowardly --diff=quick
  ```

- **Export current settings** to a YAML file (for backup or to edit and re-apply):

  ```bash
  cowardly --export=./backup.yaml
  ```

- **Reset all Brave policy settings**

  ```bash
  cowardly --reset
  cowardly -r
  ```

- **Print current settings** (user and enforced/managed when present)

  ```bash
  cowardly --view
  cowardly -v
  ```

- **List, restore, or delete backups** (plists in `~/Library/Application Support/cowardly/backups/`)

  ```bash
  cowardly --backups              # list all backup paths
  cowardly --restore=<path>        # restore user prefs from a backup (path or filename)
  cowardly --delete-backup=<path>  # delete a backup file
  cowardly --reapply              # re-apply last saved state (~/.config/cowardly/cowardly.yaml)
  cowardly --install-login-hook    # install Launch Agent to run --reapply at login
  ```

  In the TUI, use **Backups** from the main menu to list backups, then **Enter** to restore or **d** to delete (with confirmation). If settings were reverted (e.g. after restart), the main menu shows a hint and you can press **R** to re-apply your saved preset.

- **Help**
  ```bash
  cowardly --help
  cowardly -h
  ```

Before apply or reset, the **user plist is backed up** to `~/Library/Application Support/cowardly/backups/`. **Quit Brave (Cmd+Q) before resetting**—if Brave is running, macOS or Brave can rewrite the plist from cache and the reset will not stick. Brave’s in-browser “Restore settings to their original default” cannot remove **managed** policy (the plist in `/Library/Managed Preferences/`). To fully reset, use cowardly’s Reset and **approve the authentication dialog** so the managed plist is removed; then restart Brave.

**Organizational management (MDM / Intune):** If your Mac is managed by an employer or school (e.g. Microsoft Intune), they can push Brave/Chrome policies that override local settings. After a restart, the organization may re-apply its policies and your Cowardly settings can be reverted. Use **`--reapply`** to restore your desired state (saved in `~/.config/cowardly/cowardly.yaml`), or install the **login hook** (`--install-login-hook`) so Cowardly runs `--reapply` at every login. In the TUI, if the current settings differ from your saved state, a message appears and you can press **R** to re-apply. Only your IT admin can remove MDM-applied policies; see **[docs/POLICY-ENFORCEMENT.md](docs/POLICY-ENFORCEMENT.md)** for details.

## Presets

Presets are **YAML** files in [configs/presets/](configs/presets/). Add a new preset by adding a `.yaml` file there and rebuilding. See **[docs/ADDING-PRESETS.md](docs/ADDING-PRESETS.md)** for the format and instructions.

| Preset                          | Description                                                                          |
| ------------------------------- | ------------------------------------------------------------------------------------ |
| **Quick Debloat (Recommended)** | Disable telemetry, Brave Rewards/Wallet/VPN/AI/Tor, and common bloat.                |
| **Maximum Privacy**             | Blocks all telemetry, disables Brave extras, autofill, Do Not Track, plain DNS.      |
| **Balanced Privacy**            | Blocks telemetry and Brave bloat; keeps password manager; DoH automatic.             |
| **Performance Focused**         | Disable metrics and Brave Rewards/Wallet/VPN/AI; turn off background and promotions. |
| **Developer**                   | Same as above but keeps developer tools.                                             |
| **Strict Parental Controls**    | Disable incognito, force SafeSearch, disable sign-in and developer tools.            |

## Custom settings

In **Custom** mode you can toggle individual settings in four categories:

- **Telemetry & Privacy** — Metrics, Safe Browsing reporting, URL collection, feedback surveys.
- **Privacy & Security** — Safe Browsing level, autofill, password manager, sign-in, WebRTC, QUIC, cookies, Do Not Track, SafeSearch, IPFS, incognito.
- **Brave Features** — Rewards, Wallet, VPN, AI Chat, Tor, Sync.
- **Performance & Bloat** — Background mode, recommendations, shopping list, PDF externally, translate, spellcheck, promotions, search suggestions, printing, default browser prompt, developer tools.

Use **Space** to toggle, **Enter** to apply, **a** to select all, **n** to select none.

## Project layout

The repo follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout). See **[docs/PROJECT-LAYOUT.md](docs/PROJECT-LAYOUT.md)** for the directory overview, **[docs/PLATFORMS.md](docs/PLATFORMS.md)** for current and possible future platform support (macOS / Linux / Windows), and **[docs/POLICY-ENFORCEMENT.md](docs/POLICY-ENFORCEMENT.md)** for how policy enforcement works on macOS (managed vs user preferences, raw XML plist, AppleScript auth). A summary of implemented features is in **[docs/FEATURES.md](docs/FEATURES.md)**; possible future improvements are in **[docs/FUTURE.md](docs/FUTURE.md)**.

## Development

- **Go:** `make fmt` (format code, tidy modules), `make test`, `make lint` ([golangci-lint](https://golangci-lint.run) v2). See [.golangci.reference.yml](https://github.com/golangci/golangci-lint/blob/HEAD/.golangci.reference.yml) for lint config.
- **Markdown/YAML/JSON:** `make format-check` (CI check) or `make format` / `make prettier` to fix. Prettier is pinned to 3.3.2.
- **YAML:** `make lint-yaml` (yamllint; config in `.yamllint.yml`).

## Contributing

1. Fork the repo and clone it. Ensure you have **Go 1.25.6+** and (optional) **golangci-lint**, **prettier**, and **yamllint** for local checks.
2. Create a branch for your change. Make your edits; add presets in [configs/presets/](configs/presets/) if needed (see [docs/ADDING-PRESETS.md](docs/ADDING-PRESETS.md)).
3. Before opening a PR, run: `make test`, `make lint`, `make format-check`, and `make lint-yaml`. Fix any failures so CI passes.
4. Open a pull request with a short description of the change. For bugs or features, opening an issue first is welcome.

CI runs **gitleaks**, **prettier** (format check), and **yaml-lint** on push and pull requests; see [.github/workflows/README.md](.github/workflows/README.md).

## Disclaimer

This tool is not affiliated with Brave Software. It only changes macOS preference (policy) keys that Brave already supports. Use at your own risk; backup or note your settings before resetting.

## License

MIT. See [LICENSE](LICENSE).

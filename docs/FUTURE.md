# Future improvements and features

This document tracks possible enhancements, new presets, and platform work. Nothing here is committed; it is a note for contributors and maintainers.

## Presets and settings

### Default search provider

Brave/Chromium support policies for the default search engine (e.g. `DefaultSearchProviderEnabled`, `DefaultSearchProviderName`, `DefaultSearchProviderSearchURL`). Cowardly does not yet expose these in presets or the Custom TUI. Adding them would allow presets to set a non-Google (or EU-focused) default search in one click.

### Degoogle preset

A preset that minimizes Google surface in Brave:

- Reuse existing options: no sign-in, no sync, telemetry off, URL collection off.
- Add a non-Google default search (DuckDuckGo, Qwant, Ecosia, etc.) once default-search policy keys are supported.
- Optionally document or toggle Google SafeSearch behavior.

### Sovereign (EU) preset

A preset aimed at “EU-only” / sovereign defaults:

- Strong privacy (telemetry off, cookies, Do Not Track) — similar to Maximum Privacy.
- Default search set to an EU-oriented provider (e.g. Qwant, Ecosia) once supported.
- Optional: DNS or other settings pointing to EU-friendly resolvers if Brave exposes them via policy.
- Short doc or preset description clarifying that “sovereign” means EU-oriented defaults plus max privacy.

### Preset format extensions

- Support **list**-type policy values (e.g. `ExtensionInstallBlocklist`) in preset YAML if needed for new presets.
- Document any new keys in [ADDING-PRESETS.md](ADDING-PRESETS.md).

## Platform support

Cowardly is **macOS-only** today. Possible future platforms:

- **Linux** — JSON policy files under `/etc/brave/policies/` (or distro-specific path). See [PLATFORMS.md](PLATFORMS.md).
- **Windows** — Registry `BraveSoftware\Brave` and optional Group Policy. See [PLATFORMS.md](PLATFORMS.md).

Adding a platform means a new backend (e.g. under `internal/brave`) while keeping the same TUI/CLI surface.

## UX and TUI

- **Current Brave settings view** — Improve alignment/formatting of the “View current settings” output so the first entry and list are easier to scan (e.g. consistent indentation, column alignment).
- **Accessibility** — Consider keyboard-only navigation and any screen-reader hints if the TUI gains more complex views.

## Policy enforcement and docs

- **Implementation notes** — Keep [POLICY-ENFORCEMENT.md](POLICY-ENFORCEMENT.md) updated when the apply/reset flow or managed-preferences behavior changes.
- **Troubleshooting** — Add a short “Policy not applying” section to README or SETUP that points to managed vs user prefs, MDM, and the auth dialog.

## Automation and CI

- **Release automation** — Already documented in [RELEASING.md](RELEASING.md). Possible additions: release notes generation, version bump helpers.
- **Compatibility** — Optional CI job that runs against the latest Brave stable (or a fixed version) to detect policy key renames or removals.

## Contributing

If you want to work on any of these, open an issue to align with maintainers. For presets (degoogle, sovereign), add YAML under `configs/presets/` and extend the preset schema or Custom settings only if new key types are required. See [ADDING-PRESETS.md](ADDING-PRESETS.md). For supplements (e.g. new Privacy Guides–style overlays), add YAML under `configs/supplements/`; see [PRIVACY-GUIDES.md](PRIVACY-GUIDES.md) for the current supplement structure. See [Contributing](../README.md#contributing) in the main README.

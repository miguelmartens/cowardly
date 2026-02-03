# Adding custom presets

Anyone can add a new preset by adding a single **YAML** file in **configs/presets/** and rebuilding. No Go code changes are required.

## Steps

1. **Add a new file** in **configs/presets/**.

   - Use a descriptive filename. Order in the TUI is alphabetical by filename, so use a numeric prefix if you care (e.g. `07-my-preset.yaml`).
   - Only `.yaml` files in this directory are loaded.

2. **Use the preset format** (see below).

3. **Rebuild** the binary:
   ```bash
   make build
   # or
   make run
   ```

Your preset will appear in the TUI under “Apply a preset” and can be applied like any built-in preset.

## YAML format

Each file must have:

| Field         | Description                                                                              |
| ------------- | ---------------------------------------------------------------------------------------- |
| `id`          | Unique identifier (lowercase, no spaces). Used by the CLI if you ever wire a flag to it. |
| `name`        | Display name shown in the TUI.                                                           |
| `description` | One-line summary shown in the preset list.                                               |
| `settings`    | List of Brave preference entries (see below).                                            |

Each entry in `settings` must have:

| Field   | Description                                                                 |
| ------- | --------------------------------------------------------------------------- |
| `key`   | Brave policy key (same as macOS `defaults` keys under `com.brave.Browser`). |
| `value` | Value: `true`/`false` for bool, a number for integer, or a quoted string.   |
| `type`  | One of: `bool`, `integer`, `string`. Must match the value.                  |

### Example

```yaml
id: my-preset
name: My Custom Preset
description: Disable telemetry and Brave Rewards only.
settings:
  - key: MetricsReportingEnabled
    value: false
    type: bool
  - key: BraveRewardsDisabled
    value: true
    type: bool
  - key: DnsOverHttpsMode
    value: off
    type: string
```

Comments (lines starting with `#`) are allowed and ignored.

## Finding policy keys

- **From existing presets** — Look at any file in **configs/presets/** (e.g. `01-quick.yaml`, `02-max-privacy.yaml`) for keys and typical values.
- **From the TUI** — The “Custom” menu uses the same keys; see `internal/config/settings.go` for the full list and types.
- **From Brave** — Keys are the same as Chromium/Brave policy names (e.g. [Brave policy list](https://brave.com/privacy-updates/)). On macOS they are written with `defaults write com.brave.Browser <Key> <value>`.

## Common keys (reference)

| Key                                  | Type    | Example / notes           |
| ------------------------------------ | ------- | ------------------------- |
| MetricsReportingEnabled              | bool    | false                     |
| SafeBrowsingExtendedReportingEnabled | bool    | false                     |
| FeedbackSurveysEnabled               | bool    | false                     |
| BraveRewardsDisabled                 | bool    | true                      |
| BraveWalletDisabled                  | bool    | true                      |
| BraveVPNDisabled                     | bool    | true                      |
| TorDisabled                          | bool    | true                      |
| SyncDisabled                         | bool    | true                      |
| BackgroundModeEnabled                | bool    | false                     |
| DeveloperToolsDisabled               | bool    | true                      |
| BrowserSignin                        | integer | 0 = disabled              |
| IncognitoModeAvailability            | integer | 1 = disabled              |
| DnsOverHttpsMode                     | string  | "off", "automatic", etc.  |
| WebRtcIPHandling                     | string  | "disable_non_proxied_udp" |

This is a subset; copy from existing preset files or `internal/config/settings.go` for more.

## Troubleshooting

- **Preset not showing** — Ensure the file is in **configs/presets/**, has a `.yaml` extension, and parses as valid YAML. Run `make build` again.
- **Apply fails** — Check that every `type` matches the `value` (e.g. use `type: integer` and a number, not a string, for `BrowserSignin`).
- **Invalid key** — Brave will ignore unknown keys; the write may still “succeed.” Use keys from existing presets in configs/presets/ or Brave’s policy documentation.

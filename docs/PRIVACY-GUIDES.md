# Privacy Guides integration

Cowardly implements the [Privacy Guides](https://www.privacyguides.org/en/desktop-browsers/#brave) recommended Brave Browser configuration as a **supplement** that applies on top of a base preset or Custom settings.

## How it works

Privacy Guides is not a standalone preset. It adds settings that are **not** in Cowardly’s built-in presets (Quick Debloat, Maximum Privacy, etc.). When you apply Privacy Guides:

1. Choose a **base** — Quick Debloat (default), Maximum Privacy, Custom, or any other preset.
2. Cowardly merges the base settings with the Privacy Guides supplement.
3. The result is written to Brave and saved in your config.

This avoids overlap: the supplement only adds or overrides keys that presets don’t already set.

## Source and supplement file

- **Source:** [Privacy Guides — Brave](https://www.privacyguides.org/en/desktop-browsers/#brave)
- **Supplement file:** `configs/supplements/privacy-guides/recommendations.yaml`

The supplement includes:

- **Data collection** — BraveP3AEnabled, BraveStatsPingEnabled
- **Shields** — DefaultBraveAdblockSetting, DefaultBraveHttpsUpgradeSetting, DefaultBraveFingerprintingV2Setting, DefaultBraveRemember1PStorageSetting
- **Privacy & security** — BraveDeAmpEnabled, BraveDebouncingEnabled, BraveReduceLanguageEnabled, DefaultJavaScriptJitSetting

## Config format

When Privacy Guides is applied, `~/.config/cowardly/cowardly.yaml` looks like:

```yaml
preset:
  quick:
    settings: [...]
supplement:
  privacy_guides:
    settings: [...]
```

`preset.<id>.settings` is the base; `supplement.privacy_guides.settings` is the Privacy Guides overlay.

## TUI

1. Select **Privacy Guides recommendations** from the main menu.
2. If you have a prior config (preset or Custom), Cowardly uses that as the base and skips to confirmation.
3. Otherwise, choose a base preset (or **Custom** if you’ve applied Custom before).
4. Confirm with **y** / Enter to apply.

## CLI

```bash
cowardly --privacy-guides              # base from config or Quick Debloat
cowardly --privacy-guides=max-privacy  # base: Maximum Privacy
cowardly --privacy-guides=custom       # base: your saved Custom settings
```

Dry run and diff:

```bash
cowardly --dry-run=privacy-guides
cowardly --dry-run=privacy-guides:max-privacy
cowardly --diff=privacy-guides:custom
```

## Manual configuration

Some Privacy Guides recommendations have no Brave policy keys. Configure these in Brave yourself:

- Use default filter lists (Shields)
- Block Scripts (optional; disables JavaScript)
- Uncheck social media components (Shields UI)
- Automatically remove permissions from unused sites
- Use Google services for push messaging (if needed)

## Custom as base

If you apply **Custom** settings first, Privacy Guides can use them as the base. In the TUI, **Custom** appears in the base preset list when your config has `preset.custom.settings`. CLI: use `--privacy-guides=custom`.

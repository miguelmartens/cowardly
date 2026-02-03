# Platform support

Cowardly applies Brave Browser policy preferences so you can disable rewards, wallet, VPN, telemetry, and similar features. How policies are applied depends on the operating system.

## Current: macOS only

Cowardly **currently supports macOS only**. It uses:

- **User preferences:** `~/Library/Preferences/com.brave.Browser.plist` via the `defaults` command.
- **Managed preferences (enforced):** `/Library/Managed Preferences/com.brave.Browser.plist` (requires administrator privileges so Brave treats policies as mandatory).

The TUI and CLI assume macOS paths, `defaults` read/write, and (for enforced policies) the managed-preferences plist and AppleScript for the authentication dialog. See [POLICY-ENFORCEMENT.md](POLICY-ENFORCEMENT.md) for details.

## Possible future: Linux

On **Linux**, Brave (Chromium) reads policies from:

- **JSON files** under `/etc/brave/policies/` (or a similar path depending on distribution and Brave packaging), e.g. `managed_policies.json` or policy files in a `managed` subdirectory.

Adding Linux support would mean:

- Detecting Linux and Brave’s policy path on the system.
- Reading/writing JSON policy files (Chromium policy format) instead of plists.
- Using appropriate permissions (e.g. root or policy-manager group) for system-wide enforced policies.
- Keeping the same TUI/CLI surface; only the `internal/brave` (or a platform-specific backend) layer would change.

## Possible future: Windows

On **Windows**, Brave (Chromium) reads policies from:

- **Registry:** `HKLM\Software\Policies\BraveSoftware\Brave` (machine) and `HKCU\Software\Policies\BraveSoftware\Brave` (user).
- **Group Policy** (ADMX) when used in domain environments.

Adding Windows support would mean:

- Detecting Windows and reading/writing registry keys for Brave policies.
- Optionally integrating with Group Policy or documenting how cowardly-applied settings relate to GPO.
- Handling elevation (e.g. Run as administrator) for machine-level policies.
- Keeping the same TUI/CLI surface; only the platform backend would be Windows-specific.

## Summary

| Platform    | Status    | Policy mechanism (current or likely)                             |
| ----------- | --------- | ---------------------------------------------------------------- |
| **macOS**   | Supported | User + managed plist; `defaults`; AppleScript for admin.         |
| **Linux**   | Not yet   | JSON policy files under `/etc/brave/policies/` (or distro path). |
| **Windows** | Not yet   | Registry `BraveSoftware\Brave`; optional GPO.                    |

If you want to contribute Linux or Windows support, open an issue to discuss the approach and where to plug in platform-specific code (e.g. under `internal/brave` or a new `internal/brave/linux`, `internal/brave/windows`).

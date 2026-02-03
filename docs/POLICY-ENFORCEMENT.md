# Policy enforcement on macOS

This document describes why Cowardly writes to **managed preferences** when possible, and how that implementation works.

## The issue

Brave (like Chromium) supports policy keys that disable features such as Rewards, Wallet, VPN, and telemetry. On macOS, however, **where** those keys are stored determines whether Brave actually enforces them:

| Location                | Path                                                   | Enforced?                                                                  |
| ----------------------- | ------------------------------------------------------ | -------------------------------------------------------------------------- |
| **User preferences**    | `~/Library/Preferences/com.brave.Browser.plist`        | **No** — Brave may still show Rewards, Wallet, etc.                        |
| **Managed preferences** | `/Library/Managed Preferences/com.brave.Browser.plist` | **Yes** — Brave treats these as mandatory and hides/ disables the features |

If you only run `defaults write com.brave.Browser BraveRewardsDisabled -bool true`, the key is written to the user plist. After restarting Brave, the UI can still show Rewards and Wallet. To get **enforced** behavior (features hidden/disabled), the same keys must be present in the **managed** plist under `/Library/Managed Preferences/`. That path is the standard macOS location for mandatory (MDM-style) policies.

So the “issue” is: **user preferences are not sufficient for policy enforcement**; we must write to managed preferences when the user has admin rights.

## Implementation

### 1. Try managed first, fallback to user

- **`ApplySettings`** (in `internal/brave/preferences.go`) first calls **`WriteAllToManaged`**.
- If that succeeds (user approves the admin dialog), we return and the UI reports that policies were applied to the **enforced** location.
- If it fails (user cancels, or no admin rights), we fall back to **`WriteAll`**, which uses `defaults write` to the user domain. The UI then reports that settings were applied to **user preferences** and may not be enforced.

### 2. Raw XML plist for managed preferences

We do **not** use `defaults write` to build the managed plist. Instead we:

- **Generate a plist as raw XML** with a fixed header (XML declaration, DOCTYPE, `<plist version="1.0"><dict>`) so the format matches what Brave and guides like [hi-one](https://github.com/hi-one/hi-one/blob/main/Debloat-Brave-Browser-MacOS.md) use.
- For each setting we emit `<key>KeyName</key>` followed by `<true/>` / `<false/>`, `<integer>N</integer>`, or `<string>…</string>`, with XML escaping for key and string values.
- Write this XML to a **temporary file**, then copy that file to `/Library/Managed Preferences/com.brave.Browser.plist`.

This gives a single, predictable XML plist and avoids relying on whatever format `defaults write` would produce for a temporary path.

### 3. Administrator privileges via AppleScript

Writing under `/Library/Managed Preferences/` requires root. We use **AppleScript** `do shell script "…" with administrator privileges` so that:

- A **macOS GUI dialog** appears (password or Touch ID) instead of a terminal `sudo` prompt.
- The TUI and CLI keep working; the user does not type their password in the terminal.

The shell command we run (after the user approves) is:

```sh
mkdir -p "/Library/Managed Preferences" && \
cp /path/to/temp/com.brave.Browser.plist "/Library/Managed Preferences/com.brave.Browser.plist" && \
chown root:wheel "/Library/Managed Preferences/com.brave.Browser.plist" && \
chmod 644 "/Library/Managed Preferences/com.brave.Browser.plist"
```

- **chown root:wheel** — standard for system-managed preferences.
- **chmod 644** — readable by Brave, writable only by root (per common practice and the hi-one guide).

### 4. Reset

**Reset** removes keys from the user domain with `defaults delete com.brave.Browser`. If the managed plist exists, we also try to **remove** it via the same AppleScript-with-admin pattern so that Brave no longer enforces any of our policies. If the user cancels the dialog, the managed plist is left in place.

## Summary

| Aspect          | Choice                                                               | Reason                                                              |
| --------------- | -------------------------------------------------------------------- | ------------------------------------------------------------------- |
| **Where**       | `/Library/Managed Preferences/com.brave.Browser.plist` when possible | Required for Brave to enforce policies (hide Rewards/Wallet, etc.). |
| **Format**      | Raw XML plist                                                        | Predictable, matches Brave/hi-one expectations.                     |
| **Admin**       | AppleScript “with administrator privileges”                          | GUI auth dialog; works from TUI and CLI.                            |
| **Permissions** | `chown root:wheel`, `chmod 644`                                      | Standard for managed prefs; readable by Brave.                      |
| **Fallback**    | User preferences via `defaults write`                                | Works without admin; user is informed enforcement may not apply.    |

## References

- [Debloat Brave Browser (macOS)](https://github.com/hi-one/hi-one/blob/main/Debloat-Brave-Browser-MacOS.md) — managed plist path and format.
- [Chromium policy list](https://chromium.googlesource.com/chromium/src/+/HEAD/components/policy/resources/templates/policy_list.yaml) — policy keys and semantics (Brave follows Chromium policies).
- Apple: [Managed Preferences](https://developer.apple.com/documentation/devicemanagement/managed-preferences) — system location for mandatory configuration.

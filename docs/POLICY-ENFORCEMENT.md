# Policy enforcement on macOS

This document describes why Cowardly writes to **managed preferences** when possible, and how that implementation works. Cowardly currently supports **macOS** only; for possible future Linux and Windows support, see [PLATFORMS.md](PLATFORMS.md).

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

## Organizational management (MDM / Intune)

Your Mac may be **managed by an organization** (MDM, e.g. Microsoft Intune, Jamf, Kandji). In that case:

- The organization can push **Chrome/Brave policies** (e.g. `BraveRewardsDisabled`, `BraveWalletDisabled`) that apply as **Platform / Machine / Mandatory**.
- Those policies are enforced by the system and may **override** or **coexist with** anything Cowardly writes to the local managed plist.
- In `brave://policy`, such policies show **Source: Platform**, **Level: Mandatory**. The plist at `/Library/Managed Preferences/com.brave.Browser.plist` may be **absent** (no such file) because the MDM delivers policy through a different mechanism (e.g. configuration profiles), but Brave still applies them.

**If you see “Managed by your organization” or Rewards/Wallet stay disabled after Reset or a full reinstall:**

1. Check **System Settings → Privacy & Security → Profiles** (or **Profiles** in System Preferences). If the device is “supervised and managed by” a company (e.g. Cegeka, your employer), that management is the source.
2. Cowardly and the fresh-brave script **cannot remove** MDM-applied policies. Only your IT admin can change or remove the Brave/Chrome policy for your device.
3. On an **unmanaged** Mac, Reset and the script clear local plists and Brave returns to an unmanaged state; on a managed Mac, organizational policy continues to apply until IT changes it.

### Settings reverted after restart

On a managed Mac, the organization often re-applies its policies at login. That can **overwrite** the local managed plist, so the settings you applied with Cowardly are reverted after a restart.

**What you can do:**

1. **Re-apply** — Cowardly saves your last-applied preset (or file) to **`~/.config/cowardly/cowardly.yaml`**. Run `cowardly --reapply` after logging in to restore your desired state. You may need to approve the macOS authentication dialog again so the managed plist is written.
2. **Login hook** — Run `cowardly --install-login-hook` once. This installs a Launch Agent that runs `cowardly --reapply` at every login. Your desired state is then re-applied automatically (you may see the auth dialog at login).
3. **TUI** — When you start the TUI, it compares current settings to the saved desired state. If they differ, it shows a message and you can press **R** to re-apply from the main menu.

## Summary

| Aspect          | Choice                                                               | Reason                                                              |
| --------------- | -------------------------------------------------------------------- | ------------------------------------------------------------------- |
| **Where**       | `/Library/Managed Preferences/com.brave.Browser.plist` when possible | Required for Brave to enforce policies (hide Rewards/Wallet, etc.). |
| **Format**      | Raw XML plist                                                        | Predictable, matches Brave/hi-one expectations.                     |
| **Admin**       | AppleScript “with administrator privileges”                          | GUI auth dialog; works from TUI and CLI.                            |
| **Permissions** | `chown root:wheel`, `chmod 644`                                      | Standard for managed prefs; readable by Brave.                      |
| **Fallback**    | User preferences via `defaults write`                                | Works without admin; user is informed enforcement may not apply.    |
| **MDM**         | Organizational policies (Intune, etc.)                               | Can override local plists; only IT can remove them.                 |

## References

- [Debloat Brave Browser (macOS)](https://github.com/hi-one/hi-one/blob/main/Debloat-Brave-Browser-MacOS.md) — managed plist path and format.
- [Chromium policy list](https://chromium.googlesource.com/chromium/src/+/HEAD/components/policy/resources/templates/policy_list.yaml) — policy keys and semantics (Brave follows Chromium policies).
- Apple: [Managed Preferences](https://developer.apple.com/documentation/devicemanagement/managed-preferences) — system location for mandatory configuration.

# Troubleshooting

## “Forget me when I close this site” — re-login and cookies cleared

**Symptoms:** After restarting Brave you have to log in again and re-accept cookies. In Brave’s Shields settings, **“Forget me when I close this site”** appears **enabled** and **greyed out** (cannot be disabled).

**Cause:** Cowardly sets the policy **DefaultBraveRemember1PStorageSetting** to **2** (remember first-party storage) so Brave does not clear cookies/storage when you close a site. That policy is included in the **Privacy Guides** supplement. Brave also has a separate **feature flag** that controls whether the “Forget me when I close this site” option is available. When that flag is enabled (Brave’s default in some versions), the UI can show the option as on or locked even though the policy says “remember.”

**Fix:** Disable Brave’s feature flag so the option is removed and first-party storage is no longer cleared on close:

1. In Brave, open **`brave://flags`**.
2. Search for **`brave-forget-first-party-storage`** (or “forget first party”).
3. Set it to **Disabled** (not “Default”).
4. Restart Brave.

After that, the “Forget me when I close this site” option disappears from the UI and you should stay logged in across restarts.

**Verify policy:** In Brave, open **`brave://policy`** and search for **DefaultBraveRemember1PStorageSetting**. It should show **2**, Source **Platform**, Level **Mandatory**. If it is missing or 0, re-apply Privacy Guides (quit Brave first, then run `cowardly --privacy-guides` or `cowardly --privacy-guides=quick`) and approve the macOS authentication dialog so the policy is written to managed preferences.

**Note:** Cowardly only sets **policy** keys (e.g. DefaultBraveRemember1PStorageSetting) via the plist. The `brave-forget-first-party-storage` flag is a Brave/Chromium feature flag and is not configurable through Cowardly; disabling it in `brave://flags` is the correct way to turn off that behaviour.

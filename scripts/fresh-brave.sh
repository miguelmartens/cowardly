#!/usr/bin/env bash
# Completely remove Brave Browser: quit app, delete all data, uninstall via Homebrew.
# Use this for a full clean removal before reinstalling.
set -e

if [[ "$(uname -s)" != Darwin ]]; then
  echo "This script only runs on macOS." >&2
  exit 1
fi

USER_PLIST="$HOME/Library/Preferences/com.brave.Browser.plist"
MANAGED_PLIST="/Library/Managed Preferences/com.brave.Browser.plist"
APP_SUPPORT="$HOME/Library/Application Support/Brave Software/Brave Browser"
CACHES="$HOME/Library/Caches/Brave Software"
SAVED_STATE="$HOME/Library/Saved Application State/com.brave.Browser.savedState"
COOKIES="$HOME/Library/Cookies/com.brave.Browser.binarycookies"
APP_BUNDLE="/Applications/Brave Browser.app"
HOMEBREW_CASK="brave-browser"

while getopts "nh" opt; do
  case $opt in
    n) SKIP_BREW=1 ;;
    h)
      echo "Usage: $0 [-n] [-h]"
      echo "  -n    Do not run 'brew uninstall'; only remove data and app (use if Brave was not installed via Homebrew)"
      echo "  -h    Show this help"
      echo ""
      echo "By default: quits Brave, removes all preferences/profile/caches, uninstalls the app via Homebrew (brew uninstall --cask $HOMEBREW_CASK --zap)."
      exit 0
      ;;
    *) exit 1 ;;
  esac
done

echo "Brave — complete removal"
echo "========================"

# 1. Quit Brave
if pgrep -x "Brave Browser" >/dev/null 2>&1; then
  echo "Quitting Brave Browser..."
  osascript -e 'tell application "Brave Browser" to quit' 2>/dev/null || true
  sleep 2
  if pgrep -x "Brave Browser" >/dev/null 2>&1; then
    echo "Brave is still running. Quit it manually (Cmd+Q), then run this script again." >&2
    exit 1
  fi
  echo "Brave quit."
else
  echo "Brave is not running."
fi

# 2. Remove all user and profile data
echo ""
echo "Removing preferences and profile data..."

[[ -f "$USER_PLIST" ]] && rm -f "$USER_PLIST" && echo "  Removed: $USER_PLIST"
[[ -d "$APP_SUPPORT" ]] && rm -rf "$APP_SUPPORT" && echo "  Removed: $APP_SUPPORT"
[[ -d "$CACHES" ]] && rm -rf "$CACHES" && echo "  Removed: $CACHES"
[[ -d "$SAVED_STATE" ]] && rm -rf "$SAVED_STATE" && echo "  Removed: $SAVED_STATE"
[[ -f "$COOKIES" ]] && rm -f "$COOKIES" && echo "  Removed: $COOKIES"

# 3. Remove managed plist (requires admin)
if [[ -f "$MANAGED_PLIST" ]]; then
  echo "  Removing managed policy file (you may be prompted for your password)..."
  if osascript -e "do shell script \"rm -f '$MANAGED_PLIST'\" with administrator privileges" 2>/dev/null; then
    echo "  Removed: $MANAGED_PLIST"
  else
    echo "  Could not remove managed plist. Run: sudo rm -f \"$MANAGED_PLIST\"" >&2
  fi
fi

# 4. Uninstall via Homebrew (removes app and runs cask zap if defined)
if [[ -z "$SKIP_BREW" ]]; then
  echo ""
  if command -v brew >/dev/null 2>&1; then
    if brew list --cask "$HOMEBREW_CASK" >/dev/null 2>&1; then
      echo "Uninstalling Brave via Homebrew (brew uninstall --cask $HOMEBREW_CASK --zap)..."
      brew uninstall --cask "$HOMEBREW_CASK" --zap --force 2>/dev/null || brew uninstall --cask "$HOMEBREW_CASK" --force 2>/dev/null || true
      echo "  Homebrew uninstall done."
    else
      echo "Brave is not installed via Homebrew (cask $HOMEBREW_CASK not found)."
    fi
  else
    echo "Homebrew not found; skipping brew uninstall."
  fi
fi

# 5. If app bundle still exists (e.g. not installed by brew), remove it
if [[ -d "$APP_BUNDLE" ]]; then
  echo ""
  echo "Removing application: $APP_BUNDLE"
  rm -rf "$APP_BUNDLE"
  echo "  Removed."
fi

echo ""
echo "Done. Brave has been fully removed. Reinstall with: brew install --cask $HOMEBREW_CASK"

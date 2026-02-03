// Package brave provides access to Brave Browser preferences on macOS via the defaults command.
package brave

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Timeouts for subprocess calls to avoid hanging.
const (
	defaultsTimeout  = 30 * time.Second
	osascriptTimeout = 90 * time.Second // User may need time for auth dialog.
)

// plistXMLHeader matches the format used by hi-one so Brave reads the managed plist correctly.
const plistXMLHeader = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
`

const plistXMLFooter = `
</dict>
</plist>
`

const (
	// Domain is the macOS defaults domain for Brave Browser (user preferences).
	Domain = "com.brave.Browser"
	// ManagedPreferencesPath is the system path for mandatory policies (Brave enforces these and hides Rewards/Wallet etc.).
	ManagedPreferencesPath = "/Library/Managed Preferences/com.brave.Browser"
)

// ValueType is the plist value type for a preference.
type ValueType string

const (
	TypeBool    ValueType = "bool"
	TypeInteger ValueType = "integer"
	TypeString  ValueType = "string"
)

// Setting represents a single Brave preference key and its value.
type Setting struct {
	Key   string
	Value interface{}
	Type  ValueType
}

// settingsToPlistXML returns a full plist XML document for the given settings (for managed preferences).
func settingsToPlistXML(settings []Setting) string {
	var b strings.Builder
	b.WriteString(plistXMLHeader)
	for _, s := range settings {
		b.WriteString("\t<key>")
		b.WriteString(plistEscapeString(s.Key))
		b.WriteString("</key>\n\t")
		switch s.Type {
		case TypeBool:
			if v, ok := s.Value.(bool); ok && v {
				b.WriteString("<true/>")
			} else {
				b.WriteString("<false/>")
			}
		case TypeInteger:
			var n int
			switch v := s.Value.(type) {
			case int:
				n = v
			case int64:
				n = int(v)
			default:
				n = 0
			}
			b.WriteString(fmt.Sprintf("<integer>%d</integer>", n))
		case TypeString:
			b.WriteString("<string>")
			b.WriteString(plistEscapeString(fmt.Sprintf("%v", s.Value)))
			b.WriteString("</string>")
		default:
			b.WriteString("<string>")
			b.WriteString(plistEscapeString(fmt.Sprintf("%v", s.Value)))
			b.WriteString("</string>")
		}
		b.WriteString("\n")
	}
	b.WriteString(plistXMLFooter)
	return b.String()
}

// plistEscapeString escapes for use inside plist XML key or string elements.
func plistEscapeString(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// IsMacOS returns true if the current OS is darwin (macOS).
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// writeToPath runs `defaults write <path> <key> <type> <value>` for the given plist path (domain or absolute path).
func writeToPath(path string, s Setting) error {
	if !IsMacOS() {
		return fmt.Errorf("cowardly only supports macOS")
	}
	args := []string{"write", path, s.Key}
	switch s.Type {
	case TypeBool:
		v := "false"
		if b, ok := s.Value.(bool); ok && b {
			v = "true"
		}
		args = append(args, "-bool", v)
	case TypeInteger:
		var v string
		switch n := s.Value.(type) {
		case int:
			v = fmt.Sprintf("%d", n)
		case int64:
			v = fmt.Sprintf("%d", n)
		default:
			v = "0"
		}
		args = append(args, "-integer", v)
	case TypeString:
		args = append(args, "-string", fmt.Sprintf("%v", s.Value))
	default:
		return fmt.Errorf("unsupported type %q", s.Type)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultsTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "defaults", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("defaults write: %w", ctx.Err())
		}
		return fmt.Errorf("defaults write: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Write applies a single setting to user preferences using `defaults write`.
func Write(s Setting) error {
	return writeToPath(Domain, s)
}

// WriteAll applies multiple settings to user preferences; stops on first error.
func WriteAll(settings []Setting) error {
	for _, s := range settings {
		if err := Write(s); err != nil {
			return err
		}
	}
	return nil
}

// ApplySettings writes to managed preferences (enforced) when possible; otherwise to user prefs.
// Returns true if managed path was used (policies will be enforced; restart Brave).
func ApplySettings(settings []Setting) (managed bool, err error) {
	if err := WriteAllToManaged(settings); err == nil {
		return true, nil
	}
	if err := WriteAll(settings); err != nil {
		return false, err
	}
	return false, nil
}

// WriteAllToManaged writes settings to /Library/Managed Preferences/com.brave.Browser.plist
// as raw XML so Brave treats them as mandatory (enforced: Rewards/Wallet etc. are hidden).
// Uses AppleScript for admin privileges (GUI dialog); if it fails, the caller can fall back to WriteAll (user prefs).
func WriteAllToManaged(settings []Setting) error {
	if !IsMacOS() {
		return fmt.Errorf("cowardly only supports macOS")
	}
	tmpDir, err := os.MkdirTemp("", "cowardly")
	if err != nil {
		return fmt.Errorf("temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	src := filepath.Join(tmpDir, "com.brave.Browser.plist")
	xmlContent := settingsToPlistXML(settings)
	if err := os.WriteFile(src, []byte(xmlContent), 0600); err != nil {
		return fmt.Errorf("write temp plist: %w", err)
	}

	// Only the temp file path (src) is passed to the shell; preset data is in the plist content
	// and never interpolated into the shell command, so there is no shell injection from presets.
	// Use AppleScript "with administrator privileges" so a GUI dialog appears (password or Touch ID).
	// chmod 644 so the plist is readable by Brave (per hi-one / managed preferences practice).
	shellCmd := fmt.Sprintf("mkdir -p \"/Library/Managed Preferences\" && cp %s \"/Library/Managed Preferences/com.brave.Browser.plist\" && chown root:wheel \"/Library/Managed Preferences/com.brave.Browser.plist\" && chmod 644 \"/Library/Managed Preferences/com.brave.Browser.plist\"",
		shellSingleQuoted(src))
	script := `do shell script "` + escapeForAppleScript(shellCmd) + `" with administrator privileges`
	ctx, cancel := context.WithTimeout(context.Background(), osascriptTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("copy to managed preferences: %w", ctx.Err())
		}
		return fmt.Errorf("copy to managed preferences: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// shellSingleQuoted returns s wrapped in single quotes for the shell, escaping any ' in s.
func shellSingleQuoted(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// escapeForAppleScript escapes a string for use inside double quotes in AppleScript.
func escapeForAppleScript(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, `"`, `\"`)
}

// Read returns the current value for key from user preferences, or ("", false) if unset or error.
func Read(key string) (string, bool) {
	if !IsMacOS() {
		return "", false
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultsTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "defaults", "read", Domain, key)
	out, err := cmd.CombinedOutput()
	if err != nil || ctx.Err() != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}

// ManagedPlistExists returns true if the managed preferences plist is present.
func ManagedPlistExists() bool {
	if !IsMacOS() {
		return false
	}
	_, err := os.Stat(ManagedPreferencesPath + ".plist")
	return err == nil
}

// ReadManaged returns the value for key from the managed plist, or ("", false) if unset or error.
// Use this to show what Brave actually enforces (managed overrides user).
func ReadManaged(key string) (string, bool) {
	if !IsMacOS() || !ManagedPlistExists() {
		return "", false
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultsTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "defaults", "read", ManagedPreferencesPath, key)
	out, err := cmd.CombinedOutput()
	if err != nil || ctx.Err() != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}

// Delete removes a single key.
func Delete(key string) error {
	if !IsMacOS() {
		return fmt.Errorf("cowardly only supports macOS")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultsTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "defaults", "delete", Domain, key)
	_ = cmd.Run() // ignore error if key missing
	return nil
}

// Reset removes all keys for the Brave domain (user prefs) and the managed plist if present.
func Reset() error {
	if !IsMacOS() {
		return fmt.Errorf("cowardly only supports macOS")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultsTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "defaults", "delete", Domain)
	if out, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(out), "domain does not exist") {
			return fmt.Errorf("defaults delete: %w: %s", err, strings.TrimSpace(string(out)))
		}
	}
	// Remove managed plist if present so Brave no longer enforces policies (GUI auth dialog).
	if _, err := os.Stat(ManagedPreferencesPath + ".plist"); err == nil {
		dst := ManagedPreferencesPath + ".plist"
		script := `do shell script "rm -f ` + escapeForAppleScript(shellSingleQuoted(dst)) + `" with administrator privileges`
		ctx2, cancel2 := context.WithTimeout(context.Background(), osascriptTimeout)
		defer cancel2()
		cmd := exec.CommandContext(ctx2, "osascript", "-e", script)
		_ = cmd.Run() // ignore error if user cancels
	}
	return nil
}

// BraveAppPath is the default path to the Brave Browser application.
const BraveAppPath = "/Applications/Brave Browser.app"

// BraveInstalled checks if Brave Browser is installed at BraveAppPath.
func BraveInstalled() bool {
	if !IsMacOS() {
		return false
	}
	info, err := os.Stat(BraveAppPath)
	return err == nil && info.IsDir()
}

// BraveVersion returns the Brave Browser version string (e.g. "1.65.120") from the app bundle, or "" if unreadable.
// Useful for diagnostics; policy behavior may vary by Brave version.
func BraveVersion() string {
	if !IsMacOS() {
		return ""
	}
	infoPlist := filepath.Join(BraveAppPath, "Contents", "Info.plist")
	if _, err := os.Stat(infoPlist); err != nil {
		return ""
	}
	// defaults read expects the path without .plist extension
	domain := filepath.Join(BraveAppPath, "Contents", "Info")
	ctx, cancel := context.WithTimeout(context.Background(), defaultsTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "defaults", "read", domain, "CFBundleShortVersionString")
	out, err := cmd.CombinedOutput()
	if err != nil || ctx.Err() != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// BraveRunning returns true if the Brave Browser process is running.
// Quitting Brave before apply can ensure a clean state; this is used to warn the user.
func BraveRunning() bool {
	if !IsMacOS() {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultsTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pgrep", "-x", "Brave Browser")
	err := cmd.Run()
	return err == nil
}

// UserPreferencesPath returns the path to the user's Brave plist (~/Library/Preferences/com.brave.Browser.plist).
func UserPreferencesPath() (string, error) {
	if !IsMacOS() {
		return "", fmt.Errorf("cowardly only supports macOS")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	return filepath.Join(home, "Library", "Preferences", Domain+".plist"), nil
}

// BackupUserPlist copies the user preferences plist to ~/Library/Application Support/cowardly/backups/<timestamp>-user.plist.
// Creates the backup directory if needed. Returns the backup path, or error if the source plist does not exist or copy fails.
func BackupUserPlist() (string, error) {
	src, err := UserPreferencesPath()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(src); err != nil {
		return "", fmt.Errorf("user plist not found: %w", err)
	}
	home, _ := os.UserHomeDir()
	backupDir := filepath.Join(home, "Library", "Application Support", "cowardly", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}
	ts := time.Now().Format("2006-01-02T15-04-05")
	dst := filepath.Join(backupDir, ts+"-user.plist")
	data, err := os.ReadFile(src)
	if err != nil {
		return "", fmt.Errorf("read user plist: %w", err)
	}
	if err := os.WriteFile(dst, data, 0600); err != nil {
		return "", fmt.Errorf("write backup: %w", err)
	}
	return dst, nil
}

// DryRun returns a human-readable description of what ApplySettings would write (without writing).
func DryRun(settings []Setting) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Would apply %d setting(s). Target: managed preferences (enforced) if you approve the macOS dialog; otherwise user preferences.\n\n", len(settings)))
	for _, s := range settings {
		var val string
		switch s.Type {
		case TypeBool:
			if v, ok := s.Value.(bool); ok && v {
				val = "true"
			} else {
				val = "false"
			}
		case TypeInteger:
			val = fmt.Sprintf("%v", s.Value)
		case TypeString:
			val = fmt.Sprintf("%q", s.Value)
		default:
			val = fmt.Sprintf("%v", s.Value)
		}
		b.WriteString(fmt.Sprintf("  %s = %s\n", s.Key, val))
	}
	return strings.TrimSpace(b.String())
}

// settingValueStr returns the string form of a setting's value for comparison.
func settingValueStr(s Setting) string {
	switch s.Type {
	case TypeBool:
		if v, ok := s.Value.(bool); ok && v {
			return "1"
		}
		return "0"
	case TypeInteger:
		return fmt.Sprintf("%v", s.Value)
	case TypeString:
		return fmt.Sprintf("%v", s.Value)
	default:
		return fmt.Sprintf("%v", s.Value)
	}
}

// Diff returns a human-readable list of changes that would be made (current value -> new value).
// Only includes keys where the effective current value differs from the new value.
func Diff(settings []Setting) string {
	var b strings.Builder
	for _, s := range settings {
		current, _ := ReadManaged(s.Key)
		if current == "" {
			current, _ = Read(s.Key)
		}
		newStr := settingValueStr(s)
		if current == newStr {
			continue
		}
		if current == "" {
			current = "(not set)"
		}
		b.WriteString(fmt.Sprintf("  %s: %s -> %s\n", s.Key, current, newStr))
	}
	return strings.TrimSpace(b.String())
}

// ReadCurrent returns the effective current value for key (managed overrides user) and infers a type.
// Used for export. The value is normalized to the format we use in Setting (bool, int, or string).
func ReadCurrent(key string) (Setting, bool) {
	raw, ok := ReadManaged(key)
	if !ok {
		raw, ok = Read(key)
	}
	if !ok {
		return Setting{}, false
	}
	raw = strings.TrimSpace(raw)
	s := Setting{Key: key}
	switch strings.ToLower(raw) {
	case "1", "true", "yes":
		s.Type = TypeBool
		s.Value = true
		return s, true
	case "0", "false", "no":
		s.Type = TypeBool
		s.Value = false
		return s, true
	}
	if n, err := parseInt(raw); err == nil {
		s.Type = TypeInteger
		s.Value = n
		return s, true
	}
	s.Type = TypeString
	s.Value = raw
	return s, true
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

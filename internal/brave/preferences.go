// Package brave provides access to Brave Browser preferences on macOS via the defaults command.
package brave

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
	cmd := exec.Command("defaults", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
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

	// Use AppleScript "with administrator privileges" so a GUI dialog appears (password or Touch ID)
	// instead of terminal sudo, which doesn't work properly when run from the TUI.
	// chmod 644 so the plist is readable by Brave (per hi-one / managed preferences practice).
	shellCmd := fmt.Sprintf("mkdir -p \"/Library/Managed Preferences\" && cp %s \"/Library/Managed Preferences/com.brave.Browser.plist\" && chown root:wheel \"/Library/Managed Preferences/com.brave.Browser.plist\" && chmod 644 \"/Library/Managed Preferences/com.brave.Browser.plist\"",
		shellSingleQuoted(src))
	script := `do shell script "` + escapeForAppleScript(shellCmd) + `" with administrator privileges`
	cmd := exec.Command("osascript", "-e", script)
	if out, err := cmd.CombinedOutput(); err != nil {
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

// Read returns the current value for key, or ("", false) if unset or error.
func Read(key string) (string, bool) {
	if !IsMacOS() {
		return "", false
	}
	cmd := exec.Command("defaults", "read", Domain, key)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}

// Delete removes a single key.
func Delete(key string) error {
	if !IsMacOS() {
		return fmt.Errorf("cowardly only supports macOS")
	}
	cmd := exec.Command("defaults", "delete", Domain, key)
	_ = cmd.Run() // ignore error if key missing
	return nil
}

// Reset removes all keys for the Brave domain (user prefs) and the managed plist if present.
func Reset() error {
	if !IsMacOS() {
		return fmt.Errorf("cowardly only supports macOS")
	}
	cmd := exec.Command("defaults", "delete", Domain)
	if out, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(out), "domain does not exist") {
			return fmt.Errorf("defaults delete: %w: %s", err, strings.TrimSpace(string(out)))
		}
	}
	// Remove managed plist if present so Brave no longer enforces policies (GUI auth dialog).
	if _, err := os.Stat(ManagedPreferencesPath + ".plist"); err == nil {
		dst := ManagedPreferencesPath + ".plist"
		script := `do shell script "rm -f ` + escapeForAppleScript(shellSingleQuoted(dst)) + `" with administrator privileges`
		cmd := exec.Command("osascript", "-e", script)
		_ = cmd.Run() // ignore error if user cancels
	}
	return nil
}

// BraveInstalled checks if Brave Browser is installed in /Applications.
func BraveInstalled() bool {
	if !IsMacOS() {
		return false
	}
	info, err := os.Stat("/Applications/Brave Browser.app")
	return err == nil && info.IsDir()
}

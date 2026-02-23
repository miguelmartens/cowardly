// Command cowardly is a macOS TUI to debloat Brave Browser.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cowardly/cowardly/internal/brave"
	"github.com/cowardly/cowardly/internal/config"
	"github.com/cowardly/cowardly/internal/presets"
	"github.com/cowardly/cowardly/internal/ui"
	"github.com/cowardly/cowardly/internal/userconfig"
)

// Version is set at build time via -ldflags (e.g. -ldflags "-X main.Version=v0.2.0"). If unset, builds show "dev".
var Version = "dev"

func main() {
	if !brave.IsMacOS() {
		fmt.Fprintln(os.Stderr, "cowardly only supports macOS.")
		os.Exit(1)
	}

	args := os.Args[1:]
	for _, arg := range args {
		if strings.TrimLeft(arg, "-") == "beta" {
			brave.UseBeta(true)
			break
		}
	}
	for _, arg := range args {
		arg = strings.TrimLeft(arg, "-")
		switch {
		case arg == "help" || arg == "h":
			printUsage()
			return
		case arg == "version" || arg == "v":
			versionInfo()
			return
		case arg == "current" || arg == "c":
			current()
			return
		case arg == "reset" || arg == "r":
			reset()
			return
		case arg == "apply" || arg == "a":
			applyPreset("quick")
			return
		case strings.HasPrefix(arg, "apply="):
			applyPreset(strings.TrimPrefix(arg, "apply="))
			return
		case arg == "privacy-guides":
			applyPrivacyGuides(parsePrivacyGuidesBase("privacy-guides"))
			return
		case strings.HasPrefix(arg, "privacy-guides="):
			applyPrivacyGuides(strings.TrimPrefix(arg, "privacy-guides="))
			return
		case arg == "dry-run":
			dryRun("quick")
			return
		case strings.HasPrefix(arg, "dry-run="):
			dryRun(strings.TrimPrefix(arg, "dry-run="))
			return
		case strings.HasPrefix(arg, "diff="):
			diffPreset(strings.TrimPrefix(arg, "diff="))
			return
		case strings.HasPrefix(arg, "export="):
			exportSettings(strings.TrimPrefix(arg, "export="))
			return
		case strings.HasPrefix(arg, "apply-file="):
			applyFile(strings.TrimPrefix(arg, "apply-file="))
			return
		case arg == "reapply":
			reapply()
			return
		case arg == "install-login-hook":
			installLoginHook()
			return
		case arg == "backups" || arg == "b":
			listBackups()
			return
		case strings.HasPrefix(arg, "restore="):
			restoreBackup(strings.TrimPrefix(arg, "restore="))
			return
		case strings.HasPrefix(arg, "delete-backup="):
			deleteBackup(strings.TrimPrefix(arg, "delete-backup="))
			return
		}
	}

	if _, err := presets.AllWithError(); err != nil {
		fmt.Fprintf(os.Stderr, "Presets failed to load: %v\n", err)
		os.Exit(1)
	}

	if !brave.BraveInstalled() {
		which := "Brave Browser"
		if brave.IsBeta() {
			which = "Brave Browser Beta"
		}
		fmt.Fprintf(os.Stderr, "%s not found in /Applications. Install Brave first.\n", which)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func versionInfo() {
	fmt.Printf("Cowardly version: %s\n", Version)
	which := "Brave"
	if brave.IsBeta() {
		which = "Brave Beta"
	}
	if v := brave.BraveVersion(); v != "" {
		fmt.Printf("%s version: %s\n", which, v)
	} else {
		fmt.Printf("%s version: (not installed or unknown)\n", which)
	}
}

// parsePrivacyGuidesBase returns the base preset ID if presetID is privacy-guides form, else "".
// "privacy-guides" -> config base if set, else "quick"; "privacy-guides:max-privacy" -> "max-privacy".
func parsePrivacyGuidesBase(presetID string) string {
	if presetID == "privacy-guides" {
		base, _ := userconfig.PrivacyGuidesBaseFromConfig()
		if base != "" {
			return base
		}
		return presets.PrivacyGuidesBasePresetID
	}
	if strings.HasPrefix(presetID, "privacy-guides:") {
		return strings.TrimPrefix(presetID, "privacy-guides:")
	}
	return ""
}

func findPreset(id string) *presets.Preset {
	plist, _ := presets.AllWithError()
	if plist == nil {
		return nil
	}
	for i := range plist {
		if plist[i].ID == id {
			return &plist[i]
		}
	}
	return nil
}

func privacyGuidesSettings(baseID string) ([]brave.Setting, error) {
	if baseID == "custom" {
		desired, _ := userconfig.Read()
		if desired == nil || len(desired.Settings) == 0 {
			return nil, fmt.Errorf("no custom settings in config to use as base")
		}
		supplement, err := presets.LoadPrivacyGuides()
		if err != nil {
			return nil, err
		}
		return presets.MergeSettingsWithSupplement(desired.Settings, supplement), nil
	}
	return presets.PrivacyGuidesMerged(baseID)
}

func dryRun(presetID string) {
	var settings []brave.Setting
	if baseID := parsePrivacyGuidesBase(presetID); baseID != "" {
		var err error
		settings, err = privacyGuidesSettings(baseID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "privacy-guides: %v\n", err)
			os.Exit(1)
		}
	} else {
		p := findPreset(presetID)
		if p == nil {
			fmt.Fprintf(os.Stderr, "Preset %q not found.\n", presetID)
			os.Exit(1)
		}
		settings = p.Settings
	}
	fmt.Println(brave.DryRun(settings))
}

func diffPreset(presetID string) {
	var settings []brave.Setting
	if baseID := parsePrivacyGuidesBase(presetID); baseID != "" {
		var err error
		settings, err = privacyGuidesSettings(baseID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "privacy-guides: %v\n", err)
			os.Exit(1)
		}
	} else {
		p := findPreset(presetID)
		if p == nil {
			fmt.Fprintf(os.Stderr, "Preset %q not found.\n", presetID)
			os.Exit(1)
		}
		settings = p.Settings
	}
	diff := brave.Diff(settings)
	if diff == "" {
		fmt.Println("No changes (current values match preset).")
		return
	}
	fmt.Println("Would change:")
	fmt.Println(diff)
}

func applyPrivacyGuides(basePresetID string) {
	if basePresetID == "" {
		basePresetID = presets.PrivacyGuidesBasePresetID
	}
	if !brave.BraveInstalled() {
		which := "Brave Browser"
		if brave.IsBeta() {
			which = "Brave Browser Beta"
		}
		fmt.Fprintf(os.Stderr, "%s not found in /Applications.\n", which)
		os.Exit(1)
	}
	if brave.BraveRunning() {
		fmt.Fprintln(os.Stderr, "Warning: Brave is running. Quit Brave for a clean apply.")
	}
	settings, err := privacyGuidesSettings(basePresetID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "privacy-guides: %v\n", err)
		os.Exit(1)
	}
	if path, err := brave.BackupUserPlist(); err == nil {
		fmt.Fprintf(os.Stderr, "Backed up user plist to %s\n", path)
	}
	managed, err := brave.ApplySettings(settings)
	if err != nil {
		fmt.Fprintf(os.Stderr, "apply failed: %v\n", err)
		os.Exit(1)
	}
	if managed {
		fmt.Printf("Applied Privacy Guides recommendations (enforced). Restart Brave for changes to take effect.\n")
	} else {
		fmt.Printf("Applied Privacy Guides recommendations. Restart Brave. For enforced policies, approve the macOS authentication dialog when you run apply.\n")
	}
	fmt.Fprintf(os.Stderr, "Source: %s\n", presets.PrivacyGuidesURL)
	if err := userconfig.WritePrivacyGuides(basePresetID); err != nil {
		fmt.Fprintf(os.Stderr, "Note: could not save desired state to ~/.config/cowardly: %v\n", err)
	}
}

func applyPreset(presetID string) {
	if !brave.BraveInstalled() {
		which := "Brave Browser"
		if brave.IsBeta() {
			which = "Brave Browser Beta"
		}
		fmt.Fprintf(os.Stderr, "%s not found in /Applications.\n", which)
		os.Exit(1)
	}
	if brave.BraveRunning() {
		fmt.Fprintln(os.Stderr, "Warning: Brave is running. Quit Brave for a clean apply.")
	}
	p := findPreset(presetID)
	if p == nil {
		fmt.Fprintf(os.Stderr, "Preset %q not found. Use --current to list preset IDs from presets.\n", presetID)
		os.Exit(1)
	}
	if path, err := brave.BackupUserPlist(); err == nil {
		fmt.Fprintf(os.Stderr, "Backed up user plist to %s\n", path)
	}
	managed, err := brave.ApplySettings(p.Settings)
	if err != nil {
		fmt.Fprintf(os.Stderr, "apply failed: %v\n", err)
		os.Exit(1)
	}
	if managed {
		fmt.Printf("Applied preset %q (enforced). Restart Brave for changes to take effect.\n", p.Name)
	} else {
		fmt.Printf("Applied preset %q to user prefs. Restart Brave. For enforced policies, approve the macOS authentication dialog when you run apply.\n", p.Name)
	}
	if err := userconfig.WritePreset(presetID, p.Settings); err != nil {
		fmt.Fprintf(os.Stderr, "Note: could not save desired state to ~/.config/cowardly: %v\n", err)
	}
}

func applyFile(path string) {
	if !brave.BraveInstalled() {
		which := "Brave Browser"
		if brave.IsBeta() {
			which = "Brave Browser Beta"
		}
		fmt.Fprintf(os.Stderr, "%s not found in /Applications.\n", which)
		os.Exit(1)
	}
	if brave.BraveRunning() {
		fmt.Fprintln(os.Stderr, "Warning: Brave is running. Quit Brave for a clean apply.")
	}
	settings, err := presets.LoadSettingsFromFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Load file: %v\n", err)
		os.Exit(1)
	}
	if len(settings) == 0 {
		fmt.Fprintln(os.Stderr, "No settings in file.")
		os.Exit(1)
	}
	if backupPath, err := brave.BackupUserPlist(); err == nil {
		fmt.Fprintf(os.Stderr, "Backed up user plist to %s\n", backupPath)
	}
	managed, err := brave.ApplySettings(settings)
	if err != nil {
		fmt.Fprintf(os.Stderr, "apply failed: %v\n", err)
		os.Exit(1)
	}
	if managed {
		fmt.Printf("Applied %d setting(s) from file (enforced). Restart Brave.\n", len(settings))
	} else {
		fmt.Printf("Applied %d setting(s) from file to user prefs. Restart Brave.\n", len(settings))
	}
	if err := userconfig.WriteApplyFile(path, settings); err != nil {
		fmt.Fprintf(os.Stderr, "Note: could not save desired state to ~/.config/cowardly: %v\n", err)
	}
}

func exportSettings(path string) {
	keys := exportKeysList()
	var settings []brave.Setting
	for _, key := range keys {
		s, ok := brave.ReadCurrent(key)
		if ok {
			settings = append(settings, s)
		}
	}
	if len(settings) == 0 {
		fmt.Fprintln(os.Stderr, "No current settings to export.")
		os.Exit(1)
	}
	if err := presets.WriteSettingsToFile(path, settings); err != nil {
		fmt.Fprintf(os.Stderr, "export failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Exported %d setting(s) to %s\n", len(settings), path)
}

func exportKeysList() []string {
	seen := make(map[string]bool)
	keys := []string{
		"MetricsReportingEnabled", "SafeBrowsingExtendedReportingEnabled",
		"UrlKeyedAnonymizedDataCollectionEnabled", "FeedbackSurveysEnabled",
		"BraveRewardsDisabled", "BraveWalletDisabled", "BraveVPNDisabled",
		"BraveAIChatEnabled", "TorDisabled", "SyncDisabled",
		"ShoppingListEnabled", "AlwaysOpenPdfExternally", "TranslateEnabled",
		"SpellcheckEnabled", "PromotionsEnabled", "DnsOverHttpsMode",
	}
	for _, k := range keys {
		seen[k] = true
	}
	for _, cs := range config.CustomSettings() {
		if !seen[cs.Key] {
			seen[cs.Key] = true
			keys = append(keys, cs.Key)
		}
	}
	return keys
}

func reset() {
	if brave.BraveRunning() {
		fmt.Fprintln(os.Stderr, "Brave is running. Quit Brave (Cmd+Q), then run reset again. If Brave is running, it can restore the plist from memory and the reset will not stick.")
		os.Exit(1)
	}
	if path, err := brave.BackupUserPlist(); err == nil {
		fmt.Fprintf(os.Stderr, "Backed up user plist to %s\n", path)
	}
	hadManaged, managedRemoved, err := brave.Reset()
	if err != nil {
		fmt.Fprintf(os.Stderr, "reset failed: %v\n", err)
		os.Exit(1)
	}
	if !hadManaged {
		fmt.Println("User preferences cleared. No managed policy file was present, so no authentication was needed. Restart Brave.")
	} else if managedRemoved {
		fmt.Println("All Brave policy settings reset (including managed). Restart Brave.")
	} else {
		fmt.Println("User preferences cleared. The managed policy file could not be removed (did you cancel the authentication?). Run reset again and approve the dialog.")
	}
}

func current() {
	keys := []string{
		"MetricsReportingEnabled", "SafeBrowsingExtendedReportingEnabled",
		"UrlKeyedAnonymizedDataCollectionEnabled", "FeedbackSurveysEnabled",
		"BraveRewardsDisabled", "BraveWalletDisabled", "BraveVPNDisabled",
		"BraveAIChatEnabled", "TorDisabled", "SyncDisabled",
		"ShoppingListEnabled", "AlwaysOpenPdfExternally", "TranslateEnabled",
		"SpellcheckEnabled", "PromotionsEnabled", "DnsOverHttpsMode",
	}
	if brave.ManagedPlistExists() {
		fmt.Println("(Managed plist present — enforced values shown when set)")
	}
	for _, key := range keys {
		managedVal, managedOK := brave.ReadManaged(key)
		userVal, userOK := brave.Read(key)
		if managedOK {
			fmt.Printf("  %s = %s (enforced)\n", key, managedVal)
		} else if userOK {
			fmt.Printf("  %s = %s (user)\n", key, userVal)
		} else {
			fmt.Printf("  %s = (not set)\n", key)
		}
	}
}

func listBackups() {
	paths, err := brave.ListBackups()
	if err != nil {
		fmt.Fprintf(os.Stderr, "list backups: %v\n", err)
		os.Exit(1)
	}
	if len(paths) == 0 {
		fmt.Println("No backups. Apply a preset or reset to create one.")
		return
	}
	for _, p := range paths {
		fmt.Println(p)
	}
}

func restoreBackup(path string) {
	path = resolveBackupPath(path)
	if path == "" {
		fmt.Fprintln(os.Stderr, "Backup not found. Use --backups to list paths.")
		os.Exit(1)
	}
	if brave.BraveRunning() {
		fmt.Fprintln(os.Stderr, "Warning: Brave is running. Quit Brave for a clean restore.")
	}
	if err := brave.RestoreFromBackup(path); err != nil {
		fmt.Fprintf(os.Stderr, "restore failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Restored backup. Restart Brave for changes to take effect.")
}

func deleteBackup(path string) {
	path = resolveBackupPath(path)
	if path == "" {
		fmt.Fprintln(os.Stderr, "Backup not found. Use --backups to list paths.")
		os.Exit(1)
	}
	if err := brave.DeleteBackup(path); err != nil {
		fmt.Fprintf(os.Stderr, "delete failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Backup deleted.")
}

func reapply() {
	if !brave.BraveInstalled() {
		which := "Brave Browser"
		if brave.IsBeta() {
			which = "Brave Browser Beta"
		}
		fmt.Fprintf(os.Stderr, "%s not found in /Applications.\n", which)
		os.Exit(1)
	}
	if brave.BraveRunning() {
		fmt.Fprintln(os.Stderr, "Warning: Brave is running. Quit Brave for a clean apply.")
	}
	desired, err := userconfig.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "reapply: %v\n", err)
		os.Exit(1)
	}
	if desired == nil || len(desired.Settings) == 0 {
		fmt.Fprintln(os.Stderr, "No desired state saved. Apply a preset or use --apply-file first; then --reapply will restore it after a restart.")
		os.Exit(1)
	}
	if path, err := brave.BackupUserPlist(); err == nil {
		fmt.Fprintf(os.Stderr, "Backed up user plist to %s\n", path)
	}
	managed, err := brave.ApplySettings(desired.Settings)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reapply failed: %v\n", err)
		os.Exit(1)
	}
	if desired.Preset != "" {
		fmt.Printf("Re-applied preset %q. Restart Brave for changes to take effect.\n", desired.Preset)
	} else if desired.ApplyFile != "" {
		fmt.Printf("Re-applied %d setting(s) from saved config. Restart Brave.\n", len(desired.Settings))
	} else {
		fmt.Printf("Re-applied %d setting(s). Restart Brave.\n", len(desired.Settings))
	}
	if managed {
		fmt.Println("(Enforced.)")
	} else {
		fmt.Println("(User prefs; approve the macOS dialog when you run apply for enforced policies.)")
	}
}

func installLoginHook() {
	dir, err := userconfig.ConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "install-login-hook: %v\n", err)
		os.Exit(1)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "install-login-hook: %v\n", err)
		os.Exit(1)
	}
	launchAgentDir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(launchAgentDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "install-login-hook: %v\n", err)
		os.Exit(1)
	}
	cowardlyPath, err := os.Executable()
	if err != nil {
		cowardlyPath = "cowardly" // fallback to PATH
	}
	reapplyArgs := []string{"--reapply"}
	if brave.IsBeta() {
		reapplyArgs = []string{"--beta", "--reapply"}
	}
	plistPath := filepath.Join(launchAgentDir, "com.cowardly.reapply.plist")
	programArgs := append([]string{cowardlyPath}, reapplyArgs...)
	programArgsXML := ""
	for _, a := range programArgs {
		programArgsXML += fmt.Sprintf("    <string>%s</string>\n", escapePlistString(a))
	}
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.cowardly.reapply</string>
  <key>ProgramArguments</key>
  <array>
%s  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>StandardErrorPath</key>
  <string>%s/reapply.log</string>
  <key>StandardOutPath</key>
  <string>%s/reapply.log</string>
</dict>
</plist>
`, programArgsXML, dir, dir)
	if err := os.WriteFile(plistPath, []byte(plist), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "install-login-hook: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Installed Launch Agent at %s\n", plistPath)
	fmt.Println("Cowardly will run `cowardly --reapply` at login. To re-apply to managed preferences you may need to approve the macOS dialog when you log in.")
	fmt.Println("To remove: rm", plistPath)
}

// escapePlistString escapes a string for use inside a plist <string> element.
func escapePlistString(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// resolveBackupPath returns the full path if path is a filename matching a backup, or path if it's already a full path that exists.
func resolveBackupPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	paths, err := brave.ListBackups()
	if err != nil {
		return ""
	}
	for _, p := range paths {
		if p == path || strings.HasSuffix(p, path) || strings.HasSuffix(p, "/"+path) {
			return p
		}
	}
	// path might be full path that wasn't in list (e.g. stale); still allow if file exists
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func printUsage() {
	fmt.Println(`cowardly — Brave Browser debloater for macOS

Usage:
  cowardly                        Start the TUI
  cowardly --beta                 Target Brave Browser Beta (use with any command)
  cowardly --apply, -a             Apply Quick Debloat preset and exit
  cowardly --apply=<id>            Apply preset by ID (e.g. quick, max-privacy)
  cowardly --privacy-guides [=base] Apply Privacy Guides supplement (default base: quick)
  cowardly --apply-file=<path>    Apply settings from a YAML file
  cowardly --reapply              Re-apply last saved desired state (~/.config/cowardly)
  cowardly --install-login-hook    Install Launch Agent to run --reapply at login
  cowardly --dry-run [=<id>]       Show what would be applied (default: quick)
  cowardly --diff=<id>             Show which keys would change (current -> preset)
  cowardly --export=<path>         Export current settings to YAML file
  cowardly --reset, -r             Reset all Brave policy settings and exit
  cowardly --version, -v          Print cowardly and Brave version and exit
  cowardly --current, -c          Print current settings and exit
  cowardly --backups, -b           List all backup plist paths
  cowardly --restore=<path>        Restore user prefs from a backup (path or filename)
  cowardly --delete-backup=<path>  Delete a backup file
  cowardly --help, -h              Show this help

Use --beta to target Brave Browser Beta instead of stable. Restart Brave after applying or resetting settings.`)
}

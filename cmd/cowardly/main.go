// Command cowardly is a macOS TUI to debloat Brave Browser.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cowardly/cowardly/internal/brave"
	"github.com/cowardly/cowardly/internal/config"
	"github.com/cowardly/cowardly/internal/presets"
	"github.com/cowardly/cowardly/internal/ui"
)

func main() {
	if !brave.IsMacOS() {
		fmt.Fprintln(os.Stderr, "cowardly only supports macOS.")
		os.Exit(1)
	}

	args := os.Args[1:]
	for _, arg := range args {
		arg = strings.TrimLeft(arg, "-")
		switch {
		case arg == "help" || arg == "h":
			printUsage()
			return
		case arg == "view" || arg == "v":
			view()
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
		}
	}

	if _, err := presets.AllWithError(); err != nil {
		fmt.Fprintf(os.Stderr, "Presets failed to load: %v\n", err)
		os.Exit(1)
	}

	if !brave.BraveInstalled() {
		fmt.Fprintln(os.Stderr, "Brave Browser not found in /Applications. Install Brave first.")
		os.Exit(1)
	}

	p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
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

func dryRun(presetID string) {
	p := findPreset(presetID)
	if p == nil {
		fmt.Fprintf(os.Stderr, "Preset %q not found.\n", presetID)
		os.Exit(1)
	}
	fmt.Println(brave.DryRun(p.Settings))
}

func diffPreset(presetID string) {
	p := findPreset(presetID)
	if p == nil {
		fmt.Fprintf(os.Stderr, "Preset %q not found.\n", presetID)
		os.Exit(1)
	}
	diff := brave.Diff(p.Settings)
	if diff == "" {
		fmt.Println("No changes (current values match preset).")
		return
	}
	fmt.Println("Would change:")
	fmt.Println(diff)
}

func applyPreset(presetID string) {
	if !brave.BraveInstalled() {
		fmt.Fprintln(os.Stderr, "Brave Browser not found in /Applications.")
		os.Exit(1)
	}
	if brave.BraveRunning() {
		fmt.Fprintln(os.Stderr, "Warning: Brave is running. Quit Brave for a clean apply.")
	}
	p := findPreset(presetID)
	if p == nil {
		fmt.Fprintf(os.Stderr, "Preset %q not found. Use --view to list preset IDs from presets.\n", presetID)
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
}

func applyFile(path string) {
	if !brave.BraveInstalled() {
		fmt.Fprintln(os.Stderr, "Brave Browser not found in /Applications.")
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
	if path, err := brave.BackupUserPlist(); err == nil {
		fmt.Fprintf(os.Stderr, "Backed up user plist to %s\n", path)
	}
	if err := brave.Reset(); err != nil {
		fmt.Fprintf(os.Stderr, "reset failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("All Brave policy settings reset. Restart Brave.")
	fmt.Println("If you cancelled the authentication dialog, the managed plist may still exist; run reset again and approve to remove it.")
}

func view() {
	keys := []string{
		"MetricsReportingEnabled", "SafeBrowsingExtendedReportingEnabled",
		"UrlKeyedAnonymizedDataCollectionEnabled", "FeedbackSurveysEnabled",
		"BraveRewardsDisabled", "BraveWalletDisabled", "BraveVPNDisabled",
		"BraveAIChatEnabled", "TorDisabled", "SyncDisabled",
		"ShoppingListEnabled", "AlwaysOpenPdfExternally", "TranslateEnabled",
		"SpellcheckEnabled", "PromotionsEnabled", "DnsOverHttpsMode",
	}
	if v := brave.BraveVersion(); v != "" {
		fmt.Printf("Brave version: %s\n", v)
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

func printUsage() {
	fmt.Println(`cowardly — Brave Browser debloater for macOS

Usage:
  cowardly                        Start the TUI
  cowardly --apply, -a             Apply Quick Debloat preset and exit
  cowardly --apply=<id>            Apply preset by ID (e.g. quick, max-privacy)
  cowardly --apply-file=<path>    Apply settings from a YAML file
  cowardly --dry-run [=<id>]       Show what would be applied (default: quick)
  cowardly --diff=<id>             Show which keys would change (current -> preset)
  cowardly --export=<path>         Export current settings to YAML file
  cowardly --reset, -r             Reset all Brave policy settings and exit
  cowardly --view, -v              Print current settings and exit
  cowardly --help, -h              Show this help

Restart Brave Browser after applying or resetting settings.`)
}

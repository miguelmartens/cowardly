// Command cowardly is a macOS TUI to debloat Brave Browser.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cowardly/cowardly/internal/brave"
	"github.com/cowardly/cowardly/internal/presets"
	"github.com/cowardly/cowardly/internal/ui"
)

func main() {
	if !brave.IsMacOS() {
		fmt.Fprintln(os.Stderr, "cowardly only supports macOS.")
		os.Exit(1)
	}

	args := os.Args[1:]
	if len(args) > 0 {
		switch strings.TrimLeft(args[0], "-") {
		case "apply", "a":
			applyQuick()
			return
		case "reset", "r":
			reset()
			return
		case "view", "v":
			view()
			return
		case "help", "h":
			printUsage()
			return
		}
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

func applyQuick() {
	if !brave.BraveInstalled() {
		fmt.Fprintln(os.Stderr, "Brave Browser not found in /Applications.")
		os.Exit(1)
	}
	plist := presets.All()
	var quick *presets.Preset
	for i := range plist {
		if plist[i].ID == "quick" {
			quick = &plist[i]
			break
		}
	}
	if quick == nil {
		fmt.Fprintln(os.Stderr, "Quick preset not found.")
		os.Exit(1)
	}
	managed, err := brave.ApplySettings(quick.Settings)
	if err != nil {
		fmt.Fprintf(os.Stderr, "apply failed: %v\n", err)
		os.Exit(1)
	}
	if managed {
		fmt.Println("Quick Debloat applied (enforced). Restart Brave for changes to take effect.")
	} else {
		fmt.Println("Quick Debloat applied to user prefs. Restart Brave. For enforced policies (hide Rewards/Wallet), run with sudo.")
	}
}

func reset() {
	if err := brave.Reset(); err != nil {
		fmt.Fprintf(os.Stderr, "reset failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("All Brave policy settings reset. Restart Brave.")
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
	for _, key := range keys {
		val, ok := brave.Read(key)
		if ok {
			fmt.Printf("  %s = %s\n", key, val)
		} else {
			fmt.Printf("  %s = (not set)\n", key)
		}
	}
}

func printUsage() {
	fmt.Println(`cowardly — Brave Browser debloater for macOS

Usage:
  cowardly              Start the TUI
  cowardly --apply, -a   Apply Quick Debloat preset and exit
  cowardly --reset, -r   Reset all Brave policy settings and exit
  cowardly --view, -v    Print current settings and exit
  cowardly --help, -h    Show this help

Restart Brave Browser after applying or resetting settings.`)
}

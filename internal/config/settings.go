// Package config defines all customizable Brave settings for the Custom TUI option.
package config

import "github.com/cowardly/cowardly/internal/brave"

// CustomSetting describes one toggleable Brave preference for the Custom view.
type CustomSetting struct {
	Key         string
	Label       string
	Value       interface{} // value when "enabled" (applied)
	Type        brave.ValueType
	Category    string
	DisableWord string // e.g. "Disable" or "Enable"
}

// Categories and their custom settings (aligned with slimbrave-macos / bebrave).
func CustomSettings() []CustomSetting {
	return []CustomSetting{
		// Telemetry & Privacy
		{"MetricsReportingEnabled", "Metrics Reporting", false, brave.TypeBool, "Telemetry & Privacy", "Disable"},
		{"SafeBrowsingExtendedReportingEnabled", "Safe Browsing Extended Reporting", false, brave.TypeBool, "Telemetry & Privacy", "Disable"},
		{"UrlKeyedAnonymizedDataCollectionEnabled", "URL Data Collection", false, brave.TypeBool, "Telemetry & Privacy", "Disable"},
		{"FeedbackSurveysEnabled", "Feedback Surveys", false, brave.TypeBool, "Telemetry & Privacy", "Disable"},
		// Privacy & Security
		{"SafeBrowsingProtectionLevel", "Safe Browsing", 0, brave.TypeInteger, "Privacy & Security", "Disable"},
		{"AutofillAddressEnabled", "Autofill (Addresses)", false, brave.TypeBool, "Privacy & Security", "Disable"},
		{"AutofillCreditCardEnabled", "Autofill (Credit Cards)", false, brave.TypeBool, "Privacy & Security", "Disable"},
		{"PasswordManagerEnabled", "Password Manager", false, brave.TypeBool, "Privacy & Security", "Disable"},
		{"BrowserSignin", "Browser Sign-in", 0, brave.TypeInteger, "Privacy & Security", "Disable"},
		{"WebRtcIPHandling", "WebRTC IP Leak", "disable_non_proxied_udp", brave.TypeString, "Privacy & Security", "Disable"},
		{"QuicAllowed", "QUIC Protocol", false, brave.TypeBool, "Privacy & Security", "Disable"},
		{"BlockThirdPartyCookies", "Block Third Party Cookies", true, brave.TypeBool, "Privacy & Security", "Enable"},
		{"EnableDoNotTrack", "Do Not Track", true, brave.TypeBool, "Privacy & Security", "Enable"},
		{"ForceGoogleSafeSearch", "Google SafeSearch", true, brave.TypeBool, "Privacy & Security", "Force"},
		{"IPFSEnabled", "IPFS", false, brave.TypeBool, "Privacy & Security", "Disable"},
		{"IncognitoModeAvailability", "Incognito Mode", 1, brave.TypeInteger, "Privacy & Security", "Disable"},
		// Brave Features
		{"BraveRewardsDisabled", "Brave Rewards", true, brave.TypeBool, "Brave Features", "Disable"},
		{"BraveWalletDisabled", "Brave Wallet", true, brave.TypeBool, "Brave Features", "Disable"},
		{"BraveVPNDisabled", "Brave VPN", true, brave.TypeBool, "Brave Features", "Disable"},
		{"BraveAIChatEnabled", "Brave AI Chat", false, brave.TypeBool, "Brave Features", "Disable"},
		{"TorDisabled", "Tor", true, brave.TypeBool, "Brave Features", "Disable"},
		{"SyncDisabled", "Sync", true, brave.TypeBool, "Brave Features", "Disable"},
		// Performance & Bloat
		{"BackgroundModeEnabled", "Background Mode", false, brave.TypeBool, "Performance & Bloat", "Disable"},
		{"MediaRecommendationsEnabled", "Media Recommendations", false, brave.TypeBool, "Performance & Bloat", "Disable"},
		{"ShoppingListEnabled", "Shopping List", false, brave.TypeBool, "Performance & Bloat", "Disable"},
		{"AlwaysOpenPdfExternally", "Always Open PDF Externally", true, brave.TypeBool, "Performance & Bloat", "Enable"},
		{"TranslateEnabled", "Translate", false, brave.TypeBool, "Performance & Bloat", "Disable"},
		{"SpellcheckEnabled", "Spellcheck", false, brave.TypeBool, "Performance & Bloat", "Disable"},
		{"PromotionsEnabled", "Promotions", false, brave.TypeBool, "Performance & Bloat", "Disable"},
		{"SearchSuggestEnabled", "Search Suggestions", false, brave.TypeBool, "Performance & Bloat", "Disable"},
		{"PrintingEnabled", "Printing", false, brave.TypeBool, "Performance & Bloat", "Disable"},
		{"DefaultBrowserSettingEnabled", "Default Browser Prompt", false, brave.TypeBool, "Performance & Bloat", "Disable"},
		{"DeveloperToolsDisabled", "Developer Tools", true, brave.TypeBool, "Performance & Bloat", "Disable"},
	}
}

// CustomSettingsByCategory returns settings grouped by category (category name -> slice of indices into CustomSettings()).
func CustomSettingsByCategory() map[string][]int {
	all := CustomSettings()
	catIdx := make(map[string][]int)
	for i, s := range all {
		catIdx[s.Category] = append(catIdx[s.Category], i)
	}
	return catIdx
}

// CategoryOrder defines the display order of categories.
var CategoryOrder = []string{
	"Telemetry & Privacy",
	"Privacy & Security",
	"Brave Features",
	"Performance & Bloat",
}

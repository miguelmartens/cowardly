package brave

// Variant is the Brave channel (stable or beta).
type Variant string

const (
	VariantStable Variant = "stable"
	VariantBeta   Variant = "beta"
)

// currentVariant is the selected Brave channel. Default is stable.
var currentVariant = VariantStable

// UseBeta sets whether to target Brave Beta instead of Brave stable.
// Call this before any other brave package functions (e.g. at startup from a --beta flag).
func UseBeta(beta bool) {
	if beta {
		currentVariant = VariantBeta
	} else {
		currentVariant = VariantStable
	}
}

// IsBeta returns true if the current variant is Brave Beta.
func IsBeta() bool {
	return currentVariant == VariantBeta
}

// Domain returns the macOS defaults domain for the current variant.
// Stable: com.brave.Browser, Beta: com.brave.Browser.beta.
func Domain() string {
	if currentVariant == VariantBeta {
		return "com.brave.Browser.beta"
	}
	return "com.brave.Browser"
}

// ManagedPreferencesPath returns the system path for mandatory policies (without .plist).
// Stable: /Library/Managed Preferences/com.brave.Browser
// Beta: /Library/Managed Preferences/com.brave.Browser.beta
func ManagedPreferencesPath() string {
	if currentVariant == VariantBeta {
		return "/Library/Managed Preferences/com.brave.Browser.beta"
	}
	return "/Library/Managed Preferences/com.brave.Browser"
}

// BraveAppPath returns the path to the Brave application for the current variant.
// Stable: /Applications/Brave Browser.app
// Beta: /Applications/Brave Browser Beta.app
func BraveAppPath() string {
	if currentVariant == VariantBeta {
		return "/Applications/Brave Browser Beta.app"
	}
	return "/Applications/Brave Browser.app"
}

// braveProcessName returns the process name for pgrep.
// Stable: Brave Browser, Beta: Brave Browser Beta
func braveProcessName() string {
	if currentVariant == VariantBeta {
		return "Brave Browser Beta"
	}
	return "Brave Browser"
}
